package proxy

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/northbright/iocopy"
	"github.com/rs/cors"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

const (
	// have to cut them off at some point...
	// https://www.veritas.com/support/en_US/article.100028680
	forwardTimeout = 30 * time.Minute
)

// UDP: https://github.com/1lann/udp-forward ?
// TODO: one big reason to be http/https aware, is to add cors magic :/
// forward connection into the mesh
func ForwardTCPToMesh(ctx context.Context, listenAddr, localDNSAddr, meshAddr string, wireguardNet *netstack.Net) error {
	log := logr.FromContextOrDiscard(ctx)
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", listenAddr)
	if err != nil {
		log.Error(err, "error listening")
		return err
	}
	aborted := false
	go func() {
		<-ctx.Done()
		aborted = true
		listener.Close()
	}()
	for {
		incomingConnection, err := listener.Accept()
		if err != nil {
			if aborted {
				log.Info("context canceled")
				return err
			}
			log.Error(err, "error accepting connection")
			continue
		}

		tcpCtx, cancel := context.WithTimeout(ctx, forwardTimeout)

		// No DialTimeout in wg apparently
		wgConnection, err := wireguardNet.Dial("tcp", meshAddr)
		if err != nil {
			cancel()
			incomingConnection.Close()
			log.Error(err, "error forwarding connection")
			continue
		}

		go forwardTCPConnection(tcpCtx, wgConnection, incomingConnection)
	}
}

// forward connection out of the mesh to the actual network service
func ReceiveFromMesh(ctx context.Context, listenAddr string, meshPort int, wireguardNet *netstack.Net) error {
	log := logr.FromContextOrDiscard(ctx)
	listener, err := wireguardNet.ListenTCP(&net.TCPAddr{Port: meshPort})
	if err != nil {
		return err
	}
	aborted := false
	go func() {
		<-ctx.Done()
		aborted = true
		listener.Close()
	}()

	for {
		wgConnection, err := listener.Accept()
		if err != nil {
			if aborted {
				log.Info("context canceled")
				return err
			}
			log.Error(err, "error accepting connection")
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, forwardTimeout)

		serviceConnection, err := net.DialTimeout("tcp", listenAddr, 10*time.Second)
		if err != nil {
			cancel()
			wgConnection.Close()
			log.Error(err, "error forwarding connection")
			continue
		}

		go forwardTCPConnection(ctx, serviceConnection, wgConnection)
	}
}

func connCopy(ctx context.Context, from, to net.Conn, close chan<- error) {
	// TODO: how do we cancel these io.Copys?
	//_, err := io.Copy(from, to)
	ch := iocopy.Start(ctx, from, to, 2*1024*1024, 356*24*time.Hour) // TODO: tbh, I prefer the expanded explicit version - see caddyserver :)
	var err error
	for event := range ch {
		switch ev := event.(type) {
		// case *iocopy.EventWritten:
		// 	// n bytes have been written successfully.
		// 	// Get the count of bytes.
		// 	n := ev.Written()
		// 	percent := float32(float64(n) / (float64(total) / float64(100)))
		// 	log.Printf("on EventWritten: %v/%v bytes written(%.2f%%)", n, total, percent)

		case *iocopy.EventError:
			// an error occured.
			// Get the error.
			log.Printf("on EventError: %v", ev.Err())
			err = ev.Err()
			break

		// case *iocopy.EventOK:
		// 	// IO copy succeeded.
		// 	// Get the total count of written bytes.
		// 	n := ev.Written()
		// 	percent := float32(float64(n) / (float64(total) / float64(100)))
		// 	log.Printf("on EventOK: %v/%v bytes written(%.2f%%)", n, total, percent)

		// 	// Get SHA-256 checksum of remote file.
		// 	checksum := hash.Sum(nil)
		// 	fmt.Printf("SHA-256:\n%x", checksum)
		default:
			continue
		}
	}
	close <- err
}

func forwardTCPConnection(ctx context.Context, from, to net.Conn) {
	tracer := telemetry.TracerFromContext(ctx)
	log := logr.FromContextOrDiscard(ctx)

	defer from.Close()
	defer to.Close()

	ctx, forwardTCPSpan := tracer.Start(
		ctx,
		"forwardTCPConnection",
	)
	forwardTCPSpan.SetAttributes(
		attribute.String("conn.localAddr", from.LocalAddr().String()),
		attribute.String("conn.remoteAddr", from.RemoteAddr().String()),
	)
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Error(err, "panic in forwardTCPToMesh")
			forwardTCPSpan.SetAttributes(attribute.String("error", err.Error()))
		}
		forwardTCPSpan.End()
	}()

	errCh := make(chan error)

	go connCopy(ctx, to, from, errCh)
	go connCopy(ctx, from, to, errCh)
	eofs := 0
	for {
		select {
		case err := <-errCh:
			if err != nil && err != io.EOF {
				forwardTCPSpan.SetAttributes(attribute.String("error", err.Error()))
				log.Error(err, "error copying connection")
			}
			if err == io.EOF {
				eofs += 1
				if eofs == 2 {
					return
				}
			}
		case <-ctx.Done():
			log.Error(ctx.Err(), "context cancelled in connection copy")
			forwardTCPSpan.SetAttributes(attribute.String("error", ctx.Err().Error()))
			return
		}
	}
}

// http original code
func forwardHTTPToMesh(gOpts *options.GlobalOptions, localAddr, remoteDeployAddress string, wireguardNet *netstack.Net) {
	//gOpts *options.GlobalOptions, deploymentName string, wireguardNet *netstack.Net, pDev *ProxyDevice, deploymentPort int)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var replyBytes []byte

		// TODO: check that this is a valid hostname in this workgroup...
		U := r.URL
		U.Host = remoteDeployAddress
		requestURL := "http:" + U.String()
		gOpts.VDebug().Info(".. Proxying HTTP request to wg", "local addr", localAddr, "wg addr", requestURL)

		client := http.Client{
			Transport: &http.Transport{
				DialContext: wireguardNet.DialContext,
			},
		}
		res, err := client.Get(requestURL)
		if err != nil {
			w.Header().Set("Proxy-Error", err.Error())
		} else {
			// TODO: yeah, never do this - there's real streaming ffs.
			replyBytes, err = ioutil.ReadAll(res.Body)
			if err != nil {
				w.Header().Set("ProxyReadAll-Error", err.Error())
			}
		}

		w.Write(replyBytes)
	})

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)
	go http.ListenAndServe(localAddr, handler)
}

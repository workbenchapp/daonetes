package ice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"path"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"
	"github.com/workbenchapp/worknet/daoctl/lib/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ContextKey string

const GetSignalServerContextKey ContextKey = "signalserver"

// TODO: yup, https!
func GetSignalServer(ctx context.Context) string {
	serverURLDefault := "http://signal.daonetes.org:8080"
	ctxURL := ctx.Value(GetSignalServerContextKey)
	if ctxURL == nil {
		return serverURLDefault
	}
	serverURL := ctxURL.(string)
	if serverURL == "" {
		return serverURLDefault
	}
	return serverURL
}

type SignalValues map[string]string

func push(ctx context.Context, id string, values SignalValues) error {
	log := logr.FromContextOrDiscard(ctx)
	clientTrace := &httptrace.ClientTrace{
		// GotConn: func(info httptrace.GotConnInfo) {
		// 	log.Printf("POST conn (%s) was reused: %t", info.Conn.RemoteAddr().String(), info.Reused)
		// },
	}
	traceCtx := httptrace.WithClientTrace(ctx, clientTrace)
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   20 * time.Second,
	}

	_, ok := values["time"]
	if !ok {
		// give us a timestamp, so we can discard things that are old
		values["time"] = time.Now().Format(time.RFC3339)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(values); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		traceCtx,
		"POST", GetSignalServer(ctx)+path.Join("/", "push", id),
		buf,
	)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		log.Error(err, "get failed")
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		log.Error(err, "get failed")
		return err
	}
	// resp, err := http.Post(GetSignalServer(ctx)+path.Join("/", "push", id), "application/json", buf)
	// if err != nil {
	// 	return err
	// }
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http failed")
	}
	return nil
}

func pull(ctx context.Context, id string) <-chan SignalValues {
	log := logr.FromContextOrDiscard(ctx)
	clientTrace := &httptrace.ClientTrace{
		// GotConn: func(info httptrace.GotConnInfo) {
		// 	log.Printf("GET conn (%s) was reused: %t", info.Conn.RemoteAddr().String(), info.Reused)
		// },
	}
	traceCtx := httptrace.WithClientTrace(ctx, clientTrace)
	tracer := telemetry.TracerFromContext(ctx)
	traceCtx, span := tracer.Start(traceCtx, "pull")
	defer span.End()

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   20 * time.Second,
	}

	// TODO: assert that there is only one pull called for any one ID/URL
	ch := make(chan SignalValues)
	var retry time.Duration
	go func() {
		ctx, trySpan := otel.Tracer("daoctl").Start(traceCtx, "pull-try")
		defer trySpan.End()
		faild := func() {
			trySpan.End()
			if retry < 10 {
				retry++
			}
			time.Sleep(retry * time.Second)
		}
		defer close(ch)
		for {
			req, err := http.NewRequestWithContext(traceCtx, "GET", GetSignalServer(ctx)+path.Join("/", "pull", id), nil)
			if err != nil {
				if ctx.Err() == context.Canceled {
					return
				}
				log.Error(err, "get failed")
				faild()
				continue
			}
			res, err := httpClient.Do(req)
			if err != nil {
				if ctx.Err() == context.Canceled {
					return
				}
				log.Error(err, "get failed")
				faild()
				continue
			}
			defer res.Body.Close()
			retry = time.Duration(0)
			var info SignalValues

			if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
				if err == io.EOF {
					continue
				}
				if ctx.Err() == context.Canceled {
					return
				}
				log.Error(err, "get failed")
				faild()
				continue
			}

			span.SetAttributes(attribute.String("signal-server.response", spew.Sdump(info)))

			timestampString := info["time"]
			timestamp, err := time.Parse(time.RFC3339, timestampString)
			// TODO: make the 30 seconds a parameter...
			if err != nil || time.Since(timestamp) > time.Second*30 {
				log.V(2).Info("SKIPPING, too old", "info", spew.Sdump(info))
				continue
			}

			ch <- info
		}
	}()
	return ch
}

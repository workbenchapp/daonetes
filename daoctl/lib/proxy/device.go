package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/rs/cors"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/version"
	"github.com/workbenchapp/worknet/daoctl/lib/workgroup"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TODO: this needs to not be a global...
var mux *http.ServeMux //http.NewServeMux()

func AddAPIHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if mux == nil {
		// TODO: er, this ain't goroutinesafe..
		mux = http.NewServeMux()
		initAPIHandlers()
	}
	mux.HandleFunc(pattern, handler)
}

// ListenAndServeLocalhost is used to serve workgroup device info for all hosts to the DAPP - only http://localhost:9495 is safe from cors and mixed-tls-insecure errors (except on safari)
func ListenAndServeLocalhost(ctx context.Context, addr string) {
	cLog := logr.FromContextOrDiscard(ctx)

	debugOption := ctx.Value(options.Debug).(bool)
	if debugOption {
		AddAPIHandler("/debug/pprof/", pprof.Index)
	} else {
		AddAPIHandler("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text")
			w.Write(([]byte)("restart agent with --debug to enable the pprof handler"))
		})

	}
	AddAPIHandler("/device", func(w http.ResponseWriter, r *http.Request) {
		var err error
		var replyBytes []byte
		proxyDevice := r.URL.Query().Get("proxy")
		cLog.V(2).Info("get /device", "proxy", proxyDevice)
		if proxyDevice == "localhost" || proxyDevice == "" {
			result, err := workgroup.GetDeviceInfo(ctx)
			if err != nil {
				w.Header().Set("GetDeviceInfo-Error", err.Error())
			}
			result.Version = version.GetVersionString()
			result.VersionRevision = version.GetBuildRevision()
			result.VersionDate = version.GetBuildDate()
			replyBytes, err = json.MarshalIndent(result, "", "  ")
			if err != nil {
				w.Header().Set("GetDeviceInfo-Marshal-Error", err.Error())
			}
		} else {
			// TODO: check that this is a valid hostname in this workgroup...
			// TODO: convert proxyDevice(name) to the local proxy IP (or DNS when i have it)
			pDev := GetProxyDeviceInfoByName(proxyDevice)
			if pDev == nil {
				w.Header().Set("GetDeviceInfo-Proxy-Error", proxyDevice+" deviceInfo not cached yet")
			} else {
				replyBytes, err = UpdateDeviceInfoFromMesh(ctx, pDev)
				if err != nil {
					w.Header().Set("GetDeviceInfo-Proxy-Error", err.Error())
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(replyBytes)
	})

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)
	httpServer := http.Server{
		Addr:    addr,
		Handler: otelhttp.NewHandler(handler, "daoctl"),
	}
	go func() {
		<-ctx.Done()
		// TODO: this should really be a timeOut based Shutdown...
		mux = nil
		httpServer.Close()
	}()

	go httpServer.ListenAndServe()
}

// Ask the mesh peers instead of asking the chain :/
func UpdateDeviceInfoFromMesh(ctx context.Context, pDev *ProxyDevice) (replyBytes []byte, err error) {
	//pDev := GetProxyDeviceInfoByName(name)
	//if pDev == nil {
	//	return replyBytes, fmt.Errorf("no device named %s found", name)
	//} else {
	log := logr.FromContextOrDiscard(ctx)

	U := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", pDev.ProxyAddress, 9495), // ALWAYS listen to port 9495 on the wireguard network
		Path:   "/device",
	}
	requestURL := U.String()
	log.V(2).Info("Proxying request to device", "url", requestURL)
	//res, err := http.Get(requestURL)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return replyBytes, err
	}
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		return replyBytes, err
	} else {
		replyBytes, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return replyBytes, err
		} else {
			// Update the cache we use for setting up the proxies
			workgroup.UpdateDeviceStatusInfo(ctx, replyBytes)
		}
	}

	//}
	return replyBytes, err
}

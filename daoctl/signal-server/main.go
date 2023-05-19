package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/opentelemetry-go-contrib/launcher"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TODO: I was going to use the OS level service code, add mDNS, otel, and a shared cache so we can scale to more than one
// and auth by droping any data that wasn't signed by a key that's in a workgroup
// and add data expirey, but right now, that can all be future fun

// instead, there's a ../Dockerfile.signal-server that builds it, and makes an image
// and then we run it using
// docker run --name daonetes-signal-server --it --restart always --publish 8080:8080 daonetes/signal-server:latest
// and set up DNS for it, so we can move it without needing to code a change to the agent...

type SignalValues map[string]string

var (
	res = map[string]chan SignalValues{}
	mu  sync.RWMutex
)

func main() {
	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	os.Setenv("OTEL_METRICS_ENABLED", "true")
	os.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "api.honeycomb.io:4318")
	os.Setenv("OTEL_SERVICE_NAME", "signal-server")
	os.Setenv("HONEYCOMB_API_KEY", "yourkey")

	otelShutdown, err := launcher.ConfigureOpenTelemetry()
	if err != nil {
		log.Fatalf("error setting up OTel SDK - %e", err)
	}
	defer otelShutdown()

	http.Handle("/pull/", http.StripPrefix("/pull/", pullData()))
	http.Handle("/push/", http.StripPrefix("/push/", pushData()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(":%s", port),
		otelhttp.NewHandler(http.DefaultServeMux, "signal-server"),
	))
}

func pushData() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info SignalValues
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			log.Print("json decode failed:", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		fmt.Printf("push(%s): %#v\n", r.URL.RequestURI(), info)
		mu.Lock()
		defer mu.Unlock()

		ch := res[r.URL.Path]
		if ch == nil {
			ch = make(chan SignalValues, 10)
			res[r.URL.Path] = ch
		}

		select {
		default:
		case res[r.URL.Path] <- info:
		}
	})
}

func pullData() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		ch := res[r.URL.Path]
		if ch == nil {
			ch = make(chan SignalValues, 10)
			res[r.URL.Path] = ch
		}
		mu.Unlock()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		select {
		case <-ctx.Done():
			//fmt.Printf("pull(%s): nodata\n", r.URL.RequestURI())
			http.Error(w, ``, http.StatusRequestTimeout)
			return
		case v := <-ch:
			fmt.Printf("pull(%s): returns %#v\n", r.URL.RequestURI(), v)
			w.Header().Add("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(v); err != nil {
				log.Print("json encode failed:", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
	})
}

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

// The entry point for your application.
//
// Use this function to define your main request handling logic. It could be
// used to route based on the request properties (such as method or path), send
// the request to a backend, make completely new requests, and/or generate
// synthetic responses.

func main() {
	// Log service version
	fmt.Println("FASTLY_SERVICE_VERSION:", os.Getenv("FASTLY_SERVICE_VERSION"))

	fsthttp.ServeFunc(func(ctx context.Context, w fsthttp.ResponseWriter, r *fsthttp.Request) {
		// Filter requests that have unexpected methods.
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
			w.WriteHeader(fsthttp.StatusMethodNotAllowed)
			fmt.Fprintf(w, "This method is not allowed\n")
			return
		}

		p := r.URL.Path
		r.URL.Path = fmt.Sprintf("tlindsay%s", p)
		r.URL.Scheme = "https"
		r.CacheOptions.TTL = 5

		resp, err := r.Send(ctx, "github")
		if err != nil {
			w.WriteHeader(fsthttp.StatusBadGateway)
			fmt.Fprintln(w, err.Error())
			return
		}

		b, _ := fsthttp.BackendFromName(resp.Backend)

		backendHostname := fmt.Sprintf("%s://%s", r.URL.Scheme, b.HostOverride())
		if resp.StatusCode == fsthttp.StatusOK {
			body := fmt.Sprintf(
				`<html><head><meta name="go-import" content="%s%s git %s/%s"></head></html>`,
				r.URL.Hostname(),
				p,
				backendHostname,
				resp.Request.URL.Path,
			)
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len([]byte(body))))
			log.Printf("BODY: %s", body)
			fmt.Fprintln(w, body)
		} else {
			w.Header().Reset(resp.Header)
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		}
	})
}

package main

import (
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/snailforce"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
)

func main() {
	cfg := config.MustNewConfig()
	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	mux.Handle(snailforcev1connect.NewSnailforceServiceHandler(&snailforce.SnailforceService{}))
	// Example Request
	/*
		curl http://localhost:8080/snailforce.v1.SnailforceService/Register \
			--header 'content-type: application/json' \
			--header 'accept: application/json \
			--data '{"id":"foo","value":"bar"}'
	*/
	fmt.Println("Starting Snailforce @", cfg.Snailforce.URL)
	err := http.ListenAndServe(
		cfg.Snailforce.URL.Host,
		// For gRPC clients, it's convenient to support HTTP/2 without TLS. You can
		// avoid x/net/http2 by using http.ListenAndServeTLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
	log.Fatalf("listen failed: %v", err)
}

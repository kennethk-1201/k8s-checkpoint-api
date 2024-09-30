package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
)

var client *http.Client

var CERT_PATH = os.Getenv("CERT_PATH") // "/pki/ca.crt"

func initHttpsClient() {
	caCert, err := os.ReadFile(CERT_PATH)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		fmt.Print("Failed to append CA certificate to pool")
		os.Exit(1)
	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
}

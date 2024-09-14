package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const PORT_NUMBER = "8080"

func main() {
	s := &http.Server{
		Addr: fmt.Sprintf(":%s", PORT_NUMBER),
		// Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Starting server on port %s\n", PORT_NUMBER)
	log.Fatal(s.ListenAndServe())
}

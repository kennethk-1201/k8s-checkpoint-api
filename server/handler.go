package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const PORT_NUMBER = "8080"

func createCheckpoint(w http.ResponseWriter, r *http.Request) {
	return
}

func getCheckpoint(w http.ResponseWriter, r *http.Request) {
	return
}

func deleteCheckpoint(w http.ResponseWriter, r *http.Request) {
	return
}

func initRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/checkpoint", getCheckpoint).Methods("GET")
	r.HandleFunc("/checkpoint", createCheckpoint).Methods("POST")
	r.HandleFunc("/checkpoint", deleteCheckpoint).Methods("DELETE")

	return r
}

func Init() {
	s := &http.Server{
		Handler: initRoutes(),
		Addr:    fmt.Sprintf(":%s", PORT_NUMBER),
		// Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on port %s\n", PORT_NUMBER)
	log.Fatal(s.ListenAndServe())
}

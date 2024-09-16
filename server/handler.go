package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const port = "8080"
const checkpoint_directory = "/checkpoint"

func createCheckpoint(w http.ResponseWriter, r *http.Request) {
}

func getCheckpoint(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	namespace, ok := vars["namespace"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pod, ok := vars["pod"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	container, ok := vars["container"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	path := fmt.Sprintf("%s/%s/%s/%s", checkpoint_directory, namespace, pod, container)

	fileBytes, err := os.ReadFile(path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

func deleteCheckpoint(w http.ResponseWriter, r *http.Request) {
}

func initRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/checkpoint/{namespace}/{pod}/{container}", getCheckpoint).Methods("GET")
	r.HandleFunc("/checkpoint", createCheckpoint).Methods("POST")
	r.HandleFunc("/checkpoint", deleteCheckpoint).Methods("DELETE")

	return r
}

func Init() {
	s := &http.Server{
		Handler: initRoutes(),
		Addr:    fmt.Sprintf(":%s", port),
		// Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on port %s\n", port)
	log.Fatal(s.ListenAndServe())
}

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

// Running pods on the node.
type RunningPodsResponse struct {
	Kind       string  `json:"kind"`
	ApiVersion string  `json:"apiVersion"`
	Metadata   string  `json:"metadata"`
	Items      PodList `json:"items"`
}

type PodList []Pod

type Pod struct {
	Metadata PodMetadata `json:"metadata"`
	Spec     PodSpec     `json:"spec"`
}

type PodMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PodSpec struct {
	Containers string `json:"containers"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func parseCheckpointArgs(vars map[string]string) (string, string, bool) {
	namespace, ok := vars["namespace"]
	if !ok {
		return "", "", false
	}

	pod, ok := vars["pod"]
	if !ok {
		return "", "", false
	}
	return namespace, pod, true
}

func getCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace, pod, ok := parseCheckpointArgs(vars)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("%s/checkpoint-%s_%s-%s", checkpoint_directory, pod, namespace, container)
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

func checkpointPod(w http.ResponseWriter, r *http.Request) string {
	return ""
}

func getPodSpecs() {

}

func initRoutes() http.Handler {
	r := mux.NewRouter()

	// TO DO: Add create and delete checkpoint endpoints.
	r.HandleFunc("/checkpoint/{namespace}/{pod}", getCheckpoint).Methods("GET")
	r.HandleFunc("/checkpoint/{namespace}/{pod}", checkpointPod).Methods("POST")

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

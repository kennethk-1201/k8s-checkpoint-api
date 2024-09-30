package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var PORT = os.Getenv("PORT")
var CHECKPOINT_PATH = os.Getenv("CHECKPOINT_PATH") // "/checkpoint"
var NODE_NAME = os.Getenv("NODE_NAME")
var TOKEN = os.Getenv("TOKEN")
var API_SERVER = os.Getenv("API_SERVER") // "https://kubernetes.default.svc.cluster.local"

type CheckpointRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}

type GenericResponse struct {
	Msg string `json:"msg"`
}

type CheckpointContainerResponse struct {
	Items []string `json:"items"`
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

func handleGetCheckpoint(w http.ResponseWriter, r *http.Request) {

	// vars := mux.Vars(r)
	// namespace, pod, ok := parseCheckpointArgs(vars)
	// if !ok {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// path := fmt.Sprintf("%s/checkpoint-%s_%s-%s", CHECKPOINT_PATH, pod, namespace, container)
	// fileBytes, err := os.ReadFile(path)
	// if err != nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/octet-stream")
	// w.Write(fileBytes)
}

func handleCreateCheckpoint(w http.ResponseWriter, r *http.Request) {
	var req CheckpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := checkpointPod(req.Pod, req.Namespace); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{
		Msg: fmt.Sprint("checkpoint for pod %s in namespace %s was successfully created", req.Pod, req.Namespace),
	})
}

// Checkpoint a given Pod. Should replace with custom logic in the future.
func checkpointPod(podName string, namespace string) error {
	query, err := clientSet.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		FieldSelector: "metadata.name=" + podName,
	})

	if err != nil {
		return err
	}

	// temporary solution: use first pod and first container
	pod := query.Items[0]
	container := pod.Spec.Containers[0]
	node := pod.Spec.NodeName

	resp, err := checkpointContainer(node, namespace, pod.Name, container.Name)
	if err != nil {
		return err
	}

	// rename archive to a more readable format for our use case
	newCheckpointPath := fmt.Sprintf("/checkpoints/checkpoint_%s_%s.tar", podName, namespace)
	return os.Rename(resp.Items[0], newCheckpointPath)
}

// Call kubelet API to checkpoint the given container
func checkpointContainer(node string, namespace string, pod string, container string) (*CheckpointContainerResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/nodes/%s/proxy/checkpoint/%s/%s/%s", API_SERVER, node, namespace, pod, container)
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", TOKEN))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result *CheckpointContainerResponse
	if err := json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}

	return result, nil
}

func getRoutes() http.Handler {
	r := mux.NewRouter()

	// TO DO: Add create and delete checkpoint endpoints.
	r.HandleFunc("/checkpoint/{namespace}/{pod}", handleGetCheckpoint).Methods("GET")
	r.HandleFunc("/checkpoint/{namespace}/{pod}", handleCreateCheckpoint).Methods("POST")

	return r
}

func Init() {
	initHttpsClient()
	initKubernetesClient()

	s := &http.Server{
		Handler: getRoutes(),
		Addr:    fmt.Sprintf(":%s", PORT),
		// Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on port %s\n", PORT)
	log.Fatal(s.ListenAndServe())
}

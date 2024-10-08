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

	"github.com/containers/buildah"
	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var PORT = os.Getenv("PORT")
var CHECKPOINT_PATH = os.Getenv("CHECKPOINT_PATH") // "/checkpoint"
var RESTORE_PATH = os.Getenv("RESTORE_PATH")       // "/restore"
var TOKEN = os.Getenv("TOKEN")
var API_SERVER = os.Getenv("API_SERVER") // "https://kubernetes.default.svc.cluster.local"

type CreateCheckpointRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}

type RestoreCheckpointRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}

type GenericResponse struct {
	Msg string `json:"msg"`
}

type CheckpointContainerResponse struct {
	Items []string `json:"items"`
}

// Retrieve checkpoint archive from another node
func handleRestoreCheckpoint(w http.ResponseWriter, r *http.Request) {
	var req RestoreCheckpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pod, err := getPodSpec(req.Pod, req.Namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// retrieve and store in somewhere? (eg. push to local registry)
	endpoint := fmt.Sprintf("http://%s:3030/checkpoint/%s/%s", pod.Spec.NodeName, req.Namespace, req.Pod)

	resp, err := http.Get(endpoint)

	if err != nil {
		http.Error(w, "unable to checkpoint", http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		http.Error(w, "invalid request params", http.StatusBadRequest)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "checkpoint archive does not exist", http.StatusBadRequest)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "error retrieving checkpoint", http.StatusBadRequest)
		return
	}

	restore_archive_path := fmt.Sprintf("%s/checkpoint_%s_%s.tar", RESTORE_PATH, pod, req.Namespace)
	outFile, err := os.Create(restore_archive_path)

	if err != nil {
		http.Error(w, "error creating checkpoint file", http.StatusBadRequest)
		return
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		http.Error(w, "error storing checkpoint", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{
		Msg: fmt.Sprintf("checkpoint for pod %s in namespace %s was successfully migrated to %s", req.Pod, req.Namespace, pod.Spec.NodeName),
	})
}

// Endpoint for client to retrieve a checkpoint archive (should add authentication in the future)
func handleGetCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace, pod, ok := parseCheckpointArgs(vars)
	if !ok {
		http.Error(w, "invalid request params", http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("%s/checkpoint_%s_%s", CHECKPOINT_PATH, pod, namespace)
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "checkpoint archive does not exist", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
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

func handleCreateCheckpoint(w http.ResponseWriter, r *http.Request) {
	var req CreateCheckpointRequest
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
		Msg: fmt.Sprintf("checkpoint for pod %s in namespace %s was successfully created", req.Pod, req.Namespace),
	})
}

// Checkpoint a given Pod. Should replace with custom logic in the future.
func checkpointPod(podName string, namespace string) error {
	pod, err := getPodSpec(podName, namespace)
	if err != nil {
		return err
	}

	container := pod.Spec.Containers[0]
	node := pod.Spec.NodeName

	resp, err := checkpointContainer(node, namespace, pod.Name, container.Name)
	if err != nil {
		return err
	}

	// rename archive to a more readable format for our use case
	newCheckpointPath := fmt.Sprintf("/%s/checkpoint_%s_%s.tar", CHECKPOINT_PATH, podName, namespace)
	return os.Rename(resp.Items[0], newCheckpointPath)
}

func getPodSpec(podName string, namespace string) (*corev1.Pod, error) {
	query, err := clientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + podName,
	})

	if err != nil {
		return nil, err
	}

	// temporary solution: use first pod and first container
	return &query.Items[0], nil
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
	r.HandleFunc("/retrieve/{namespace}/{pod}", handleRestoreCheckpoint).Methods("POST")

	return r
}

func Init() {
	initHttpsClient()
	initKubernetesClient()
	if buildah.InitReexec() {
		return
	}

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

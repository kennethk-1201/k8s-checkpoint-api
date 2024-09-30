package server

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientSet *kubernetes.Clientset

func initKubernetesClient() {
	kubeConfig, err := rest.InClusterConfig()
	c, err := kubernetes.NewForConfig(kubeConfig)

	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	clientSet = c
}

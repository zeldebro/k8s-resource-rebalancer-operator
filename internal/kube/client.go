package kube

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"path/filepath"
)

func ConnectCluster() (*kubernetes.Clientset, *metricsclient.Clientset, error) {
	var config *rest.Config
	var err error

	// 1 Try in-cluster config first (when running inside pod)
	config, err = rest.InClusterConfig()
	if err != nil {

		// 2 fallback to local kubeconfig (for local testing)
		home := homedir.HomeDir()
		if home == "" {
			return nil, nil, fmt.Errorf("cannot find home directory for kubeconfig")
		}

		kubeconfig := filepath.Join(home, ".kube", "config")

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed loading kubeconfig: %w", err)
		}
	}

	// 3 Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating k8s client: %w", err)
	}

	// 4 Metrics clientset
	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating metrics client: %w", err)
	}

	return clientset, metricsClient, nil
}

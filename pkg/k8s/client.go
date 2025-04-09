package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents a Kubernetes client wrapper
type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error

	// If kubeconfig is provided, use it, otherwise use in-cluster config
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
} 
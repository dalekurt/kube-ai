package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientConfig holds the configuration for a Kubernetes client
type ClientConfig struct {
	// Path to the kubeconfig file
	KubeconfigPath string
	// Kubernetes context to use
	Context string
	// Kubernetes namespace to use
	Namespace string
	// If true, operations will target all namespaces
	AllNamespaces bool
}

// Client represents a Kubernetes client wrapper
type Client struct {
	clientset kubernetes.Interface
	config    ClientConfig
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) (*Client, error) {
	return NewClientWithConfig(ClientConfig{KubeconfigPath: kubeconfig})
}

// NewClientWithConfig creates a new Kubernetes client with the given configuration
func NewClientWithConfig(config ClientConfig) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if config.KubeconfigPath != "" {
		loadingRules.ExplicitPath = config.KubeconfigPath
	}

	// Create config overrides
	overrides := &clientcmd.ConfigOverrides{}

	// Apply context override if specified
	if config.Context != "" {
		overrides.CurrentContext = config.Context
	}

	// Apply namespace override if specified
	if config.Namespace != "" {
		overrides.Context.Namespace = config.Namespace
	}

	// Create client config
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	)

	// Create rest config
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		// If we couldn't load from kubeconfig, try in-cluster config
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// If namespace wasn't explicitly provided, get it from the client config
	if config.Namespace == "" && !config.AllNamespaces {
		namespace, _, err := clientConfig.Namespace()
		if err == nil && namespace != "" {
			config.Namespace = namespace
		} else {
			// Default to "default" namespace if not specified
			config.Namespace = "default"
		}
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() kubernetes.Interface {
	return c.clientset
}

// GetNamespace returns the currently configured namespace
func (c *Client) GetNamespace() string {
	return c.config.Namespace
}

// IsAllNamespaces returns whether operations should target all namespaces
func (c *Client) IsAllNamespaces() bool {
	return c.config.AllNamespaces
}

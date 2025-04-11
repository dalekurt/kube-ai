package k8s

import (
	"github.com/spf13/cobra"
)

// AddKubectlFlags adds standard kubectl flags to a cobra command
func AddKubectlFlags(cmd *cobra.Command) {
	// Common kubectl flags that users expect
	cmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace")
	cmd.PersistentFlags().String("context", "", "Kubernetes context")
	cmd.PersistentFlags().String("kubeconfig", "", "Path to kubeconfig file")
	cmd.PersistentFlags().BoolP("all-namespaces", "A", false, "Use all namespaces")

	// Additional common kubectl flags
	cmd.PersistentFlags().String("cluster", "", "Kubernetes cluster")
	cmd.PersistentFlags().String("user", "", "Kubernetes user")
	cmd.PersistentFlags().BoolP("insecure-skip-tls-verify", "k", false, "Skip TLS verification")
	cmd.PersistentFlags().String("certificate-authority", "", "Path to a certificate authority file")
	cmd.PersistentFlags().String("server", "", "Kubernetes API server address")
	cmd.PersistentFlags().String("token", "", "Bearer token for authentication")
}

// GetClientConfigFromFlags extracts a ClientConfig from command flags
func GetClientConfigFromFlags(cmd *cobra.Command) (ClientConfig, error) {
	config := ClientConfig{}

	// Extract values from flags
	namespace, _ := cmd.Flags().GetString("namespace")
	context, _ := cmd.Flags().GetString("context")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")

	// Set the config values
	config.Namespace = namespace
	config.Context = context
	config.KubeconfigPath = kubeconfig
	config.AllNamespaces = allNamespaces

	return config, nil
}

// NewClientFromFlags creates a new Kubernetes client using configuration from command flags
func NewClientFromFlags(cmd *cobra.Command) (*Client, error) {
	config, err := GetClientConfigFromFlags(cmd)
	if err != nil {
		return nil, err
	}

	return NewClientWithConfig(config)
}

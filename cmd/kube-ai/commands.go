package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"kube-ai/internal/config"
	"kube-ai/pkg/ai"
	"kube-ai/pkg/ai/analyzers"
	"kube-ai/pkg/k8s"
	"kube-ai/pkg/k8s/logs"
)

// createRootCommand creates the root command for the kube-ai CLI
func createRootCommand(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kube-ai",
		Short: "AI-powered Kubernetes assistant",
		Long:  `Kube-AI is an AI-powered assistant for Kubernetes, providing intelligent assistance for cluster management.`,
	}

	// Add subcommands
	rootCmd.AddCommand(createAnalyzeCmd(cfg, aiService))
	rootCmd.AddCommand(createOptimizeCmd(cfg, aiService))
	rootCmd.AddCommand(createScalingCmd(cfg, aiService))
	rootCmd.AddCommand(createGenerateCmd(cfg, aiService))
	rootCmd.AddCommand(createExplainCmd(cfg, aiService))
	rootCmd.AddCommand(createVersionCmd())

	// Add log analysis command
	rootCmd.AddCommand(createAnalyzeLogsCmd(cfg, aiService))

	// Add configuration/provider management commands
	rootCmd.AddCommand(createChatCmd(cfg, aiService))
	rootCmd.AddCommand(createSetModelCmd(cfg, aiService))
	rootCmd.AddCommand(createListModelsCmd(cfg, aiService))
	rootCmd.AddCommand(createSetProviderCmd(cfg, aiService))
	rootCmd.AddCommand(createListProvidersCmd(cfg, aiService))
	rootCmd.AddCommand(createSetApiKeyCmd(cfg, aiService))

	return rootCmd
}

// createAnalyzeCmd creates the analyze command
func createAnalyzeCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var resourceName string
	var namespace string
	var filename string

	cmd := &cobra.Command{
		Use:   "analyze [resource-type] [resource-name]",
		Short: "Analyze Kubernetes resources",
		Long:  `Analyze Kubernetes resources and provide insights and recommendations.`,
		Run: func(cmd *cobra.Command, args []string) {
			var deploymentYAML string
			var err error

			if filename != "" {
				// Read from file
				data, err := os.ReadFile(filename)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				deploymentYAML = string(data)
			} else if len(args) >= 2 {
				// Get from kubernetes
				resourceType := args[0]
				resourceName = args[1]

				// Initialize the Kubernetes client but don't use it yet in this example
				_, err := k8s.NewClient(cfg.KubeConfigPath)
				if err != nil {
					log.Fatalf("Error creating Kubernetes client: %v", err)
				}

				// This is a simplified example - in a real implementation,
				// you would need to get the YAML representation of the resource
				deploymentYAML = fmt.Sprintf("Resource type: %s, name: %s, namespace: %s",
					resourceType, resourceName, namespace)
			} else {
				log.Fatalf("Please provide resource type and name or use --filename flag")
			}

			result, err := aiService.AnalyzeDeployment(deploymentYAML)
			if err != nil {
				log.Fatalf("Error analyzing deployment: %v", err)
			}

			fmt.Println(result)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML file to analyze")

	return cmd
}

// createOptimizeCmd creates the optimize command
func createOptimizeCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var namespace string
	var filename string

	cmd := &cobra.Command{
		Use:   "optimize [options]",
		Short: "Optimize resource usage",
		Long:  `Suggest optimizations for resource usage in Kubernetes deployments.`,
		Run: func(cmd *cobra.Command, args []string) {
			var resourceYAML string
			var err error

			if filename != "" {
				// Read from file
				data, err := os.ReadFile(filename)
				if err != nil {
					log.Fatalf("Error reading file: %v", err)
				}
				resourceYAML = string(data)
			} else {
				log.Fatalf("Please provide a YAML file with --filename flag")
			}

			result, err := aiService.OptimizeResources(resourceYAML)
			if err != nil {
				log.Fatalf("Error optimizing resources: %v", err)
			}

			fmt.Println(result)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML file to optimize")

	return cmd
}

// createScalingCmd creates the scaling command
func createScalingCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var resourceName string
	var namespace string
	var metricsFile string
	var configFile string

	cmd := &cobra.Command{
		Use:   "suggest-scaling [resource-name]",
		Short: "Suggest scaling strategies",
		Long:  `Suggest optimal scaling strategies for Kubernetes workloads.`,
		Run: func(cmd *cobra.Command, args []string) {
			var metricsData string
			var configData string
			var err error

			if len(args) > 0 {
				resourceName = args[0]
			}

			if metricsFile != "" {
				data, err := os.ReadFile(metricsFile)
				if err != nil {
					log.Fatalf("Error reading metrics file: %v", err)
				}
				metricsData = string(data)
			} else {
				metricsData = "No metrics data provided."
			}

			if configFile != "" {
				data, err := os.ReadFile(configFile)
				if err != nil {
					log.Fatalf("Error reading config file: %v", err)
				}
				configData = string(data)
			} else if resourceName != "" {
				// In a real implementation, you would get the current configuration from Kubernetes
				configData = fmt.Sprintf("Resource: %s, Namespace: %s", resourceName, namespace)
			} else {
				log.Fatalf("Please provide a resource name or configuration file")
			}

			result, err := aiService.SuggestScalingStrategy(metricsData, configData)
			if err != nil {
				log.Fatalf("Error suggesting scaling strategy: %v", err)
			}

			fmt.Println(result)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVarP(&metricsFile, "metrics", "m", "", "File containing metrics data")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "File containing current configuration")

	return cmd
}

// createGenerateCmd creates the generate command
func createGenerateCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var descriptionFile string

	cmd := &cobra.Command{
		Use:   "generate [description]",
		Short: "Generate Kubernetes manifests",
		Long:  `Generate Kubernetes manifests from descriptions.`,
		Run: func(cmd *cobra.Command, args []string) {
			var description string
			var err error

			if descriptionFile != "" {
				data, err := os.ReadFile(descriptionFile)
				if err != nil {
					log.Fatalf("Error reading description file: %v", err)
				}
				description = string(data)
			} else if len(args) > 0 {
				description = strings.Join(args, " ")
			} else {
				log.Fatalf("Please provide a description or a description file")
			}

			result, err := aiService.GenerateManifest(description)
			if err != nil {
				log.Fatalf("Error generating manifest: %v", err)
			}

			fmt.Println(result)
		},
	}

	cmd.Flags().StringVarP(&descriptionFile, "file", "f", "", "File containing manifest description")

	return cmd
}

// createExplainCmd creates the explain command
func createExplainCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var errorFile string

	cmd := &cobra.Command{
		Use:   "explain [error-message]",
		Short: "Explain Kubernetes errors",
		Long:  `Explain Kubernetes errors in simple terms and suggest fixes.`,
		Run: func(cmd *cobra.Command, args []string) {
			var errorMessage string
			var err error

			if errorFile != "" {
				data, err := os.ReadFile(errorFile)
				if err != nil {
					log.Fatalf("Error reading error file: %v", err)
				}
				errorMessage = string(data)
			} else if len(args) > 0 {
				errorMessage = strings.Join(args, " ")
			} else {
				// Try to read from stdin
				stdinData, err := io.ReadAll(os.Stdin)
				if err != nil || len(stdinData) == 0 {
					log.Fatalf("Please provide an error message or use --file flag")
				}
				errorMessage = string(stdinData)
			}

			result, err := aiService.ExplainError(errorMessage)
			if err != nil {
				log.Fatalf("Error explaining Kubernetes error: %v", err)
			}

			fmt.Println(result)
		},
	}

	cmd.Flags().StringVarP(&errorFile, "file", "f", "", "File containing error message")

	return cmd
}

// createChatCmd creates the chat command
func createChatCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [message]",
		Short: "Chat about Kubernetes",
		Long:  `Have a conversation about Kubernetes topics.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalf("Please provide a message to chat about")
			}

			message := strings.Join(args, " ")
			result, err := aiService.Chat(message)
			if err != nil {
				log.Fatalf("Error in chat: %v", err)
			}

			fmt.Println(result)
		},
	}

	return cmd
}

// createSetModelCmd creates the set-model command
func createSetModelCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-model [model-name]",
		Short: "Set the default AI model",
		Long:  `Set the default AI model to use for kube-ai commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalf("Please provide a model name")
			}

			modelName := args[0]
			aiService.SetModelName(modelName)

			fmt.Printf("Model set to: %s\n", modelName)
			fmt.Println("Note: This is a temporary setting that will reset when the tool exits.")
			fmt.Printf("To make this permanent, set the %s_DEFAULT_MODEL environment variable.\n",
				strings.ToUpper(aiService.GetCurrentProvider()))
		},
	}

	return cmd
}

// createListModelsCmd creates the list-models command
func createListModelsCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-models",
		Short: "List available AI models",
		Long:  `List available AI models from Ollama.`,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := aiService.ListModels()
			if err != nil {
				log.Fatalf("Error listing models: %v", err)
			}

			fmt.Println(result)
		},
	}

	return cmd
}

// createSetProviderCmd creates the set-provider command
func createSetProviderCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-provider [provider-name]",
		Short: "Set the AI provider",
		Long:  `Set the AI provider to use for kube-ai commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalf("Please provide a provider name")
			}

			providerName := strings.ToLower(args[0])
			err := aiService.SwitchProvider(providerName)
			if err != nil {
				log.Fatalf("Error switching provider: %v", err)
			}

			// Check if API key is required but not set
			provider := aiService.GetProvider()
			if provider.RequiresAPIKey() && len(cfg.GetAPIKey(providerName)) == 0 {
				fmt.Printf("Warning: Provider '%s' requires an API key, but none is set.\n", providerName)
				fmt.Printf("Use 'kubectl ai set-api-key [provider] [key]' to set the API key.\n")
			}

			fmt.Printf("Provider set to: %s\n", providerName)
			fmt.Println("Note: This is a temporary setting that will reset when the tool exits.")
			fmt.Println("To make this permanent, set the AI_PROVIDER environment variable.")
		},
	}

	return cmd
}

// createListProvidersCmd creates the list-providers command
func createListProvidersCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-providers",
		Short: "List available AI providers",
		Long:  `List available AI providers that can be used with kube-ai.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(aiService.ListProviders())
			fmt.Printf("\nCurrent provider: %s\n", aiService.GetCurrentProvider())
			fmt.Printf("Current model: %s\n", aiService.GetCurrentModel())
			fmt.Println("\nTo change provider, use 'kubectl ai set-provider [provider-name]'")
		},
	}

	return cmd
}

// createSetApiKeyCmd creates the set-api-key command
func createSetApiKeyCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var setGlobal bool

	cmd := &cobra.Command{
		Use:   "set-api-key [provider] [api-key]",
		Short: "Set the API key for an AI provider",
		Long:  `Set the API key for an AI provider. This key will be used for authentication with the provider's API.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				log.Fatalf("Please provide both provider name and API key")
			}

			providerName := strings.ToLower(args[0])
			apiKey := args[1]

			// Verify provider is valid
			validProvider := false
			for _, provider := range []string{"openai", "anthropic", "gemini"} {
				if providerName == provider {
					validProvider = true
					break
				}
			}

			if !validProvider {
				log.Fatalf("Unsupported provider for API key: %s", providerName)
			}

			// Set the API key and save configuration
			switch providerName {
			case "openai":
				cfg.OpenAIApiKey = apiKey
			case "anthropic":
				cfg.AnthropicApiKey = apiKey
			case "gemini":
				cfg.GeminiApiKey = apiKey
			}

			// Save the configuration
			if err := cfg.SaveConfig(); err != nil {
				fmt.Printf("Warning: Failed to save configuration: %v\n", err)
			}

			fmt.Printf("API key for %s has been set.\n", providerName)

			if setGlobal {
				fmt.Printf("To make this permanent, set the %s_API_KEY environment variable.\n",
					strings.ToUpper(providerName))
			}

			// If this is the current provider, update it
			if providerName == aiService.GetCurrentProvider() {
				err := aiService.SwitchProvider(providerName)
				if err != nil {
					log.Fatalf("Error updating provider with new API key: %v", err)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&setGlobal, "global", "g", false, "Show instructions for setting API key globally")

	return cmd
}

// createAnalyzeLogsCmd creates the analyze-logs command
func createAnalyzeLogsCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var namespace string
	var resourceType string
	var resourceName string
	var container string
	var tailLines int64
	var sinceSeconds int64
	var previous bool
	var errorsOnly bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "analyze-logs [resource-type] [resource-name]",
		Short: "Analyze logs from a Kubernetes resource using AI",
		Long: `Analyze logs from a Kubernetes resource (pod, deployment, etc.) and provide 
AI-powered troubleshooting insights, including potential issues and solutions.`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Extract arguments
			resourceType = args[0]
			resourceName = args[1]

			// Create Kubernetes client
			client, err := k8s.NewClient(cfg.KubeConfigPath)
			if err != nil {
				log.Fatalf("Error creating Kubernetes client: %v", err)
			}

			// Create log collector
			collector := logs.NewLogCollector(client.GetClientset())

			// Prepare log options
			var tl *int64
			if tailLines > 0 {
				tl = &tailLines
			}

			var ss *int64
			if sinceSeconds > 0 {
				ss = &sinceSeconds
			}

			options := logs.LogOptions{
				ResourceType: resourceType,
				ResourceName: resourceName,
				Namespace:    namespace,
				Container:    container,
				TailLines:    tl,
				SinceSeconds: ss,
				Previous:     previous,
				Follow:       false,
			}

			// Collect logs
			fmt.Printf("Collecting logs from %s/%s in namespace %s...\n", resourceType, resourceName, namespace)
			logEntries, err := collector.GetResourceLogs(context.Background(), options)
			if err != nil {
				log.Fatalf("Error collecting logs: %v", err)
			}

			fmt.Printf("Collected %d log entries\n", len(logEntries))

			// Parse and analyze logs
			fmt.Println("Analyzing logs...")
			logSummary := logs.ParseLogs(logEntries)

			// Create log analyzer
			analyzer := analyzers.NewLogAnalyzer(aiService)

			// Perform analysis
			var analysisResult *analyzers.LogAnalysisResult
			if errorsOnly {
				analysisResult, err = analyzer.AnalyzeErrorLogs(context.Background(), logEntries)
			} else {
				analysisResult, err = analyzer.AnalyzeLogs(context.Background(), logEntries, logSummary)
			}

			if err != nil {
				log.Fatalf("Error analyzing logs: %v", err)
			}

			// Display results based on output format
			switch outputFormat {
			case "json":
				displayJSONResults(logSummary, analysisResult)
			default:
				displayFormattedResults(logSummary, analysisResult)
			}
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the resource")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name (for pods with multiple containers)")
	cmd.Flags().Int64VarP(&tailLines, "tail", "t", 1000, "Number of lines to include from the end of the logs")
	cmd.Flags().Int64VarP(&sinceSeconds, "since", "s", 3600, "Only return logs newer than a duration in seconds")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "Include logs from previously terminated containers")
	cmd.Flags().BoolVarP(&errorsOnly, "errors-only", "e", false, "Analyze only error logs")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text or json)")

	return cmd
}

// displayJSONResults outputs analysis results in JSON format
func displayJSONResults(summary logs.LogSummary, analysis *analyzers.LogAnalysisResult) {
	// Combine summary and analysis into a single structure
	result := struct {
		Summary  logs.LogSummary             `json:"summary"`
		Analysis analyzers.LogAnalysisResult `json:"analysis"`
	}{
		Summary:  summary,
		Analysis: *analysis,
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting JSON output: %v", err)
	}

	fmt.Println(string(jsonData))
}

// displayFormattedResults outputs analysis results in human-readable format
func displayFormattedResults(summary logs.LogSummary, analysis *analyzers.LogAnalysisResult) {
	// Determine severity color
	var severityColor string
	switch analysis.Severity {
	case "Critical":
		severityColor = "\033[1;31m" // Bold Red
	case "High":
		severityColor = "\033[31m" // Red
	case "Medium":
		severityColor = "\033[33m" // Yellow
	case "Low":
		severityColor = "\033[32m" // Green
	default:
		severityColor = "\033[0m" // Default
	}

	resetColor := "\033[0m"

	// Display log summary
	fmt.Println("\n====== LOG SUMMARY ======")
	fmt.Printf("Total Entries: %d (%d errors, %d warnings)\n",
		summary.TotalEntries, summary.ErrorCount, summary.WarningCount)
	fmt.Printf("Time Range: %s to %s (%s)\n",
		summary.TimeRange.Start.Format(time.RFC3339),
		summary.TimeRange.End.Format(time.RFC3339),
		summary.TimeRange.Duration.String())

	// Display error hotspots
	if len(summary.ErrorHotspots) > 0 {
		fmt.Println("\n=== Error Hotspots ===")
		for _, hotspot := range summary.ErrorHotspots {
			fmt.Printf("- %s: %d errors\n", hotspot.ResourceName, hotspot.ErrorCount)
		}
	}

	// Display analysis results
	fmt.Println("\n====== AI ANALYSIS ======")
	fmt.Printf("Severity: %s%s%s\n\n", severityColor, analysis.Severity, resetColor)

	fmt.Println("=== Summary ===")
	fmt.Println(analysis.Summary)

	fmt.Println("\n=== Root Causes ===")
	for i, cause := range analysis.RootCauses {
		fmt.Printf("%d. %s\n", i+1, cause)
	}

	fmt.Println("\n=== Recommended Solutions ===")
	for i, solution := range analysis.Solutions {
		fmt.Printf("%d. %s\n", i+1, solution)
	}

	if len(analysis.AdditionalInfo) > 0 {
		fmt.Println("\n=== Additional Information ===")
		for i, info := range analysis.AdditionalInfo {
			fmt.Printf("%d. %s\n", i+1, info)
		}
	}
}

// createVersionCmd creates the version command
func createVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version, git commit, and build information for kube-ai.`,
		Run: func(cmd *cobra.Command, args []string) {
			version := "dev"
			commit := "none"
			buildDate := "unknown"

			fmt.Printf("Kube-AI - Kubernetes AI Tool\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", buildDate)
		},
	}

	return cmd
}

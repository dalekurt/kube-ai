package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"kube-ai/internal/config"
	"kube-ai/pkg/ai"
	"kube-ai/pkg/ai/analyzers"
	"kube-ai/pkg/k8s"
	"kube-ai/pkg/k8s/logs"
	"kube-ai/pkg/version"
)

// createRootCommand creates the root command for the kube-ai CLI
func createRootCommand(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kube-ai",
		Short: "AI-powered Kubernetes assistant",
		Long:  `Kube-AI is an AI-powered assistant for Kubernetes, providing intelligent assistance for cluster management.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Update kubeconfig path in cfg if set via flag
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
			if kubeconfig != "" {
				cfg.KubeConfigPath = kubeconfig
			}
		},
	}

	// Add standard kubectl flags to all commands
	k8s.AddKubectlFlags(rootCmd)

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

	// Add persona command
	rootCmd.AddCommand(createPersonaCmd(cfg))

	return rootCmd
}

// createAnalyzeCmd creates the analyze command
func createAnalyzeCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
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
				resourceName := args[1]

				// Initialize the Kubernetes client with kubectl flags
				client, err := k8s.NewClientFromFlags(cmd)
				if err != nil {
					log.Fatalf("Error creating Kubernetes client: %v", err)
				}

				// Get the namespace from the client (which respects kubectl flags)
				namespace := client.GetNamespace()

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

	// Add command-specific flags (filename is not a standard kubectl flag)
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML file to analyze")

	return cmd
}

// createOptimizeCmd creates the optimize command
func createOptimizeCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
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

	// Add command-specific flags
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "YAML file to optimize")

	return cmd
}

// createScalingCmd creates the scaling command
func createScalingCmd(cfg *config.Config, aiService *ai.Service) *cobra.Command {
	var metricsFile string
	var configFile string

	cmd := &cobra.Command{
		Use:   "suggest-scaling [resource-name]",
		Short: "Suggest scaling strategies",
		Long:  `Suggest optimal scaling strategies for Kubernetes workloads.`,
		Run: func(cmd *cobra.Command, args []string) {
			var resourceName string
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
				// Initialize the Kubernetes client with kubectl flags
				client, err := k8s.NewClientFromFlags(cmd)
				if err != nil {
					log.Fatalf("Error creating Kubernetes client: %v", err)
				}

				// Get the namespace from the client (which respects kubectl flags)
				namespace := client.GetNamespace()

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

	// Add command-specific flags
	cmd.Flags().StringVarP(&metricsFile, "metrics", "m", "", "Metrics data file")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Current configuration file")

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
		Long:  `List available AI models from the current AI provider.`,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := aiService.ListModels()
			if err != nil {
				log.Fatalf("Error listing models: %v", err)
			}

			// Add information about current provider and model
			var formattedOutput strings.Builder
			formattedOutput.WriteString(result)
			formattedOutput.WriteString("\n")
			formattedOutput.WriteString(fmt.Sprintf("Current provider: %s\n", aiService.GetCurrentProvider()))
			formattedOutput.WriteString(fmt.Sprintf("Current model: %s\n", aiService.GetCurrentModel()))

			fmt.Println(formattedOutput.String())
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
	var container string
	var tailLines int64
	var sinceSeconds int64
	var previous bool
	var errorsOnly bool
	var outputFormat string
	var showLogs bool = true // Default to showing logs
	var maxLogs int = 20     // Default to 20 logs
	var tailLiveLogs bool    // New flag for live log tailing

	cmd := &cobra.Command{
		Use:   "analyze-logs [resource-type] [resource-name]",
		Short: "Analyze logs from a Kubernetes resource using AI",
		Long: `Analyze logs from a Kubernetes resource (pod, deployment, etc.) and provide 
AI-powered troubleshooting insights, including potential issues and solutions.

By default, the command will display the first 20 log entries being analyzed.
Use --show-logs=false to hide logs or --max-logs to change the number of logs shown.
Use --tail to continuously stream logs in real-time instead of analyzing a fixed set.`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Extract arguments
			resourceType := args[0]
			resourceName := args[1]

			// Create Kubernetes client with kubectl flags
			client, err := k8s.NewClientFromFlags(cmd)
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

			// Get namespace from client which respects kubectl flags
			namespace := client.GetNamespace()

			options := logs.LogOptions{
				ResourceType: resourceType,
				ResourceName: resourceName,
				Namespace:    namespace,
				Container:    container,
				TailLines:    tl,
				SinceSeconds: ss,
				Previous:     previous,
				Follow:       tailLiveLogs,
			}

			// Collect logs
			fmt.Printf("Collecting logs from %s/%s in namespace %s...\n", resourceType, resourceName, namespace)

			// Handle live tailing mode differently
			if tailLiveLogs {
				fmt.Println("Streaming logs in real-time (press Ctrl+C to stop)...")

				// Create context that can be canceled on interrupt
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Setup signal handling for graceful exit
				interruptChan := make(chan os.Signal, 1)
				signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

				// Start a goroutine that will cancel the context when interrupted
				go func() {
					<-interruptChan
					fmt.Println("\nInterrupted, stopping log stream...")
					cancel()
				}()

				// Stream logs in real-time
				logChan := make(chan logs.LogEntry)
				errChan := make(chan error)

				go func() {
					err := collector.StreamLogs(ctx, options, logChan, errChan)
					if err != nil {
						fmt.Printf("Error streaming logs: %v\n", err)
					}
				}()

				// Process streamed logs
				for {
					select {
					case entry, ok := <-logChan:
						if !ok {
							return
						}
						displayLogEntry(entry)
					case err, ok := <-errChan:
						if !ok {
							return
						}
						fmt.Printf("Error: %v\n", err)
					case <-ctx.Done():
						return
					}
				}
			}

			// Normal log collection and analysis mode
			logEntries, err := collector.GetResourceLogs(context.Background(), options)
			if err != nil {
				log.Fatalf("Error collecting logs: %v", err)
			}

			fmt.Printf("Collected %d log entries\n", len(logEntries))

			// Display logs if requested
			if showLogs {
				logCount := len(logEntries)
				if maxLogs > 0 && maxLogs < logCount {
					logCount = maxLogs
				}

				fmt.Printf("\n====== LOG ENTRIES ======\n")
				fmt.Printf("Showing %d of %d log entries:\n\n", logCount, len(logEntries))

				for i, entry := range logEntries {
					if i >= logCount {
						break
					}

					displayLogEntry(entry)
				}

				if len(logEntries) > logCount {
					fmt.Printf("\n... and %d more log entries\n", len(logEntries)-logCount)
				}
				fmt.Println()
			}

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

	// Add command-specific flags (not available in standard kubectl)
	cmd.Flags().StringVarP(&container, "container", "c", "", "Container name for pods with multiple containers")
	cmd.Flags().Int64VarP(&tailLines, "tail", "t", 1000, "Number of lines to include from the end of logs")
	cmd.Flags().Int64VarP(&sinceSeconds, "since", "s", 3600, "Only return logs newer than a duration in seconds")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "Include logs from previously terminated containers")
	cmd.Flags().BoolVarP(&errorsOnly, "errors-only", "e", false, "Analyze only error logs")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text or json)")
	cmd.Flags().BoolVar(&showLogs, "show-logs", true, "Display log entries being analyzed")
	cmd.Flags().IntVar(&maxLogs, "max-logs", 20, "Maximum number of logs to display")
	cmd.Flags().BoolVar(&tailLiveLogs, "live", false, "Stream logs in real-time")

	return cmd
}

// displayLogEntry formats and displays a single log entry with color coding
func displayLogEntry(entry logs.LogEntry) {
	// Format timestamp for readability
	timeStr := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Add colors based on log level
	levelColor := ""
	resetColor := "\033[0m"

	switch entry.LogLevel {
	case "ERROR", "FATAL":
		levelColor = "\033[31m" // Red
	case "WARN", "WARNING":
		levelColor = "\033[33m" // Yellow
	case "INFO":
		levelColor = "\033[32m" // Green
	}

	// Print log entry with container name if available
	containerInfo := ""
	if entry.ContainerName != "" {
		containerInfo = fmt.Sprintf(" [%s]", entry.ContainerName)
	}

	fmt.Printf("%s [%s%s%s]%s %s\n",
		timeStr,
		levelColor,
		entry.LogLevel,
		resetColor,
		containerInfo,
		entry.Content)
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
			fmt.Printf("Kube-AI - Kubernetes AI Tool\n")
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.GitCommit)
			fmt.Printf("Built: %s\n", version.BuildDate)
		},
	}

	return cmd
}

// createPersonaCmd creates a command for managing AI personas
func createPersonaCmd(cfg *config.Config) *cobra.Command {
	personaCmd := &cobra.Command{
		Use:   "persona",
		Short: "Manage AI assistant personas",
		Long:  "Manage different personalities for the AI assistant, affecting how it responds to queries",
	}

	// List available personas
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available personas",
		Long:  "Display all available personas, including default and custom ones",
		Run: func(cmd *cobra.Command, args []string) {
			personas := cfg.ListPersonas()

			fmt.Println("Available personas:")
			fmt.Println("-------------------")

			for name, persona := range personas {
				activeMarker := " "
				if name == cfg.ActivePersona {
					activeMarker = "*"
				}

				fmt.Printf("[%s] %s: %s\n", activeMarker, name, persona.Description)
			}

			fmt.Println("\n* = currently active persona")
		},
	}

	// Use a specific persona
	useCmd := &cobra.Command{
		Use:   "use [persona-name]",
		Short: "Set the active persona",
		Long:  "Change the active persona used by the AI assistant",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			personaName := args[0]

			err := cfg.SetPersona(personaName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			fmt.Printf("Persona changed to: %s\n", personaName)
		},
	}

	// Add a custom persona
	addCmd := &cobra.Command{
		Use:   "add [name] [description] [system-prompt]",
		Short: "Add a custom persona",
		Long:  "Create a new custom persona with a specific system prompt",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			description := args[1]
			systemPrompt := args[2]

			err := cfg.AddCustomPersona(name, description, systemPrompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			fmt.Printf("Added new persona: %s\n", name)
		},
	}

	// Remove a custom persona
	removeCmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a custom persona",
		Long:  "Delete a custom persona (default personas cannot be removed)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			err := cfg.RemoveCustomPersona(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			fmt.Printf("Removed persona: %s\n", name)
		},
	}

	// Add subcommands to persona command
	personaCmd.AddCommand(listCmd)
	personaCmd.AddCommand(useCmd)
	personaCmd.AddCommand(addCmd)
	personaCmd.AddCommand(removeCmd)

	return personaCmd
}

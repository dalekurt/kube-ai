package main

import (
	"fmt"
	"log"
	"os"

	"kube-ai/internal/config"
	"kube-ai/pkg/ai"
	"kube-ai/pkg/version"
)

func main() {
	fmt.Printf("Kube-AI - Kubernetes AI Tool (version: %s, commit: %s, built at: %s)\n",
		version.Version, version.GitCommit, version.BuildDate)

	// Load configuration
	cfg := config.LoadConfig()

	// Create AI service
	aiService := ai.NewService(cfg)

	// Create and execute the root command
	rootCmd := createRootCommand(cfg, aiService)
	if err := rootCmd.Execute(); err != nil {
		log.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}

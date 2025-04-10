package main

import (
	"log"
	"os"

	"kube-ai/internal/config"
	"kube-ai/pkg/ai"
)

func main() {
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

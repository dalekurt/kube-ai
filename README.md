# Kube-AI
<img width="1840" alt="Screenshot 2025-04-10 at 11 27 44 AM" src="https://github.com/user-attachments/assets/0102523a-d1c4-4c41-85c4-5fc91fd48c50" />
Kube-AI is an AI-powered tool for Kubernetes that helps automate and enhance Kubernetes operations, providing intelligent assistance for cluster management and application deployment.

## Features


- **Resource Analysis**: Analyze Kubernetes resources for best practices and potential issues
- **Resource Optimization**: Get AI-powered recommendations for optimizing CPU and memory usage
- **Scaling Strategies**: Receive intelligent scaling suggestions based on workload patterns
- **Manifest Generation**: Generate Kubernetes manifests from natural language descriptions
- **Error Explanation**: Get AI-powered explanations and solutions for Kubernetes errors
- **Log Analysis**: Analyze Kubernetes logs to identify patterns, issues, and solutions
- **Multi-Provider Support**: Works with multiple AI providers (Ollama, OpenAI, Anthropic, Gemini, AnythingLLM)
- **Customizable**: Switch between providers and models based on your needs

## Installation

### Using Krew

If you have [Krew](https://krew.sigs.k8s.io/) installed, you can easily install kube-ai:

```bash
kubectl krew install ai
```

### Manual Installation

If you prefer to install manually:

1. Download the appropriate version for your operating system and architecture from the [releases page](https://github.com/dalekurt/kube-ai/releases).
2. Rename the downloaded file to `kubectl-ai`.
3. Make it executable: `chmod +x kubectl-ai`
4. Move it to a directory in your PATH, e.g., `mv kubectl-ai /usr/local/bin/`

### Testing plugin installation locally

 `kubectl krew install --manifest=ai.yaml --archive=foo.tar.gz`

### Generate sha256

 `shasum -a 256 releases/download/<VERSION>/*.tar.gz`

## Usage

Once installed, you can use kube-ai with the following commands:

### Resource Analysis

Analyze Kubernetes resources for best practices and potential issues:

```bash
# Analyze a deployment
kubectl ai analyze deployment my-deployment

# Analyze from a YAML file
kubectl ai analyze -f deployment.yaml
```

### Resource Optimization

Get AI-powered recommendations for optimizing CPU and memory usage:

```bash
# Optimize resources for a deployment file
kubectl ai optimize -f deployment.yaml
```

### Scaling Strategies

Receive intelligent scaling suggestions based on workload patterns:

```bash
# Get scaling recommendations
kubectl ai suggest-scaling my-deployment

# With metrics data
kubectl ai suggest-scaling -m metrics-data.json -c current-config.yaml
```

### Manifest Generation

Generate Kubernetes manifests from natural language descriptions:

```bash
# Generate a manifest from a description
kubectl ai generate "Create a stateful MySQL database with 5GB of persistent storage"

# Generate from a description file
kubectl ai generate -f description.txt
```

### Error Explanation

Get AI-powered explanations and solutions for Kubernetes errors:

```bash
# Explain an error message
kubectl ai explain "Failed to pull image: ErrImagePull"

# Pipe kubectl error to explanation
kubectl get pods 2>&1 | kubectl ai explain
```

### Log Analysis

Analyze Kubernetes logs with AI to identify issues and get troubleshooting recommendations:

```bash
# Analyze logs from a deployment
kubectl ai analyze-logs deployment my-app -n my-namespace

# Analyze logs from a specific pod
kubectl ai analyze-logs pod my-app-pod-1234 -n my-namespace

# Analyze only error logs
kubectl ai analyze-logs deployment my-app --errors-only

# Get JSON output for further processing
kubectl ai analyze-logs deployment my-app --output json > analysis.json
```

Available options:
- `--namespace, -n`: Namespace of the resource (default: "default")
- `--container, -c`: Container name for pods with multiple containers
- `--tail, -t`: Number of lines to include from the end of logs (default: 1000)
- `--since, -s`: Only return logs newer than a duration in seconds (default: 3600)
- `--previous, -p`: Include logs from previously terminated containers
- `--errors-only, -e`: Analyze only error logs
- `--output, -o`: Output format (text or json) 

### AI Provider Management

Kube-AI supports multiple AI providers:

- **Ollama** (default, local): Uses a locally running Ollama instance
- **OpenAI**: Uses OpenAI GPT models via API
- **Anthropic**: Uses Anthropic Claude models via API  
- **Gemini**: Uses Google's Gemini models via API
- **AnythingLLM**: Uses a locally running AnythingLLM instance

#### List Available Providers

To see all available providers and which one is currently active:

```bash
kubectl ai list-providers
```

#### Switch Between Providers

To change the active AI provider:

```bash
kubectl ai set-provider [provider-name]
```

Example:
```bash
kubectl ai set-provider openai
```

#### Set API Key

For providers that require an API key (OpenAI, Anthropic, Gemini):

```bash
kubectl ai set-api-key [provider] [api-key]
```

Example:
```bash
kubectl ai set-api-key openai sk-your-api-key
kubectl ai set-api-key anthropic sk-ant-your-api-key
kubectl ai set-api-key gemini your-gemini-api-key
```

### Model Management

#### List Available Models

To see all available models for the current provider:

```bash
kubectl ai list-models
```

#### Set Default Model

To change the model used by the current provider:

```bash
kubectl ai set-model [model-name]
```

Example:
```bash
kubectl ai set-model gpt-4
kubectl ai set-model llama3.3
kubectl ai set-model claude-3-opus-20240229
```

## Configuration

Kube-AI stores its configuration in `~/.kube-ai/config.json`. This includes:

- The active AI provider
- API keys for different providers
- Default models
- Provider URLs

You can configure environment variables to set defaults:

- `AI_PROVIDER`: Default AI provider (e.g., "ollama", "openai")
- `OPENAI_API_KEY`: API key for OpenAI
- `ANTHROPIC_API_KEY`: API key for Anthropic
- `GEMINI_API_KEY`: API key for Gemini
- `OLLAMA_URL`: URL for Ollama (default: http://localhost:11434)
- `ANYTHINGLLM_URL`: URL for AnythingLLM (default: http://localhost:3001)
- `OLLAMA_DEFAULT_MODEL`: Default model for Ollama (default: llama3.3)
- `OPENAI_DEFAULT_MODEL`: Default model for OpenAI (default: gpt-3.5-turbo)
- `ANTHROPIC_DEFAULT_MODEL`: Default model for Anthropic (default: claude-3-haiku-20240307)
- `GEMINI_DEFAULT_MODEL`: Default model for Gemini (default: gemini-1.5-pro)

## Project Structure

```
kube-ai/
├── cmd/             # Application entry points
├── pkg/             # Public packages
│   ├── k8s/         # Kubernetes client utilities
│   │   └── logs/    # Kubernetes log collection and parsing
│   ├── ai/          # AI service and integration
│   │   ├── providers/  # AI provider implementations
│   │   └── analyzers/  # Specialized analyzers (logs, etc.)
│   └── version/     # Version information
├── internal/        # Private packages
│   ├── auth/        # Authentication utilities
│   └── config/      # Configuration management
├── scripts/         # Helper scripts for release and development
```

## Development

This project uses Go modules for dependency management.

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/dalekurt/kube-ai.git
cd kube-ai

# Install dependencies
go mod download

# Build the binary
go build -o kube-ai ./cmd/kube-ai

# Run tests
go test -v ./...
```

### Adding a New Provider

To add a new AI provider:

1. Create a new file in `pkg/ai/providers/` that implements the `Provider` interface
2. Update the provider factory in `pkg/ai/providers/factory.go`
3. Add provider constants and configuration in `internal/config/config.go`

### Using Taskfile

This project uses Taskfile for common development tasks:

```bash
# Install task (https://taskfile.dev)
go install github.com/go-task/task/v3/cmd/task@latest

# Build for your platform
task build

# Build for all platforms
task build:all

# Run tests
task test

# Create a release (creates a tag and pushes it)
task release -- 1.0.0
```

### Release Process

The project includes scripts to automate the release process:

#### Update CHANGELOG.md

The `update-changelog.sh` script automatically generates an updated CHANGELOG.md entry based on git commits:

```bash
# Automatically calculate next version and update CHANGELOG.md
./scripts/update-changelog.sh

# Specify a specific version
./scripts/update-changelog.sh 1.0.0
```

The script categorizes commits based on conventional commit format:
- `feat:` or `add:` prefixes are categorized as "Added"
- `fix:` or `bug:` prefixes are categorized as "Fixed"
- `change:`, `refactor:`, or `chore:` prefixes are categorized as "Changed"

#### Create GitHub Release

The `auto-release.sh` script automates the GitHub release process:

```bash
# Create a release for version 0.1.0
./scripts/auto-release.sh 0.1.0
```

This script:
1. Validates that the version follows semantic versioning format
2. Checks that the version exists in CHANGELOG.md
3. Extracts release notes from CHANGELOG.md
4. Builds binaries for all platforms
5. Creates a GitHub release draft with the binaries and release notes
6. Builds and pushes a Docker image

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT License](LICENSE)

## Support

If you encounter any issues or have questions, please file an issue on the [GitHub repository](https://github.com/dalekurt/kube-ai/issues).

## Security Note

Kube-AI stores API keys in the configuration file. In a production environment, you may want to implement more secure key storage methods or use environment variables for sensitive information.

## Security

### Supply Chain Security

This project implements [SLSA Level 3](https://slsa.dev) supply chain security using the SLSA GitHub Actions workflow. 
This provides the following security guarantees:

- **Build Provenance**: Cryptographic verification of how and where the software was built
- **Source Integrity**: Verification that the source code hasn't been tampered with
- **Build Integrity**: Protection against tampering during the build process
- **Common Vulnerabilities**: Automated scanning for known vulnerabilities

When downloading releases, you can verify their SLSA provenance to ensure they were built securely through our GitHub Actions workflows. 

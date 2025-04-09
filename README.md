# Kube-AI

Kube-AI is an AI-powered tool for Kubernetes that helps automate and enhance Kubernetes operations, providing intelligent assistance for cluster management and application deployment.

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

```bash
kubectl krew install --manifest=ai.yaml --archive=<generated-archive>.tar.gz
```

### Generate sha256

```bash
shasum -a 256 releases/download/<VERSION>/*.tar.gz
```

## Usage

Once installed, you can use kube-ai with the following commands:

### Execute AI operations

```bash
kubectl ai <command> [options]
```

For example:

```bash
kubectl ai analyze deployment my-app
kubectl ai optimize resource-usage
kubectl ai suggest scaling-strategy
```

### AI Provider Management

Kube-AI now supports multiple AI providers:

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

### Chat with the AI

Start a conversation about Kubernetes:

```bash
kubectl ai chat "How do I implement a StatefulSet?"
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
│   ├── ai/          # AI service and integration
│   │   └── providers/  # AI provider implementations
├── internal/        # Private packages
│   ├── auth/        # Authentication utilities
│   └── config/      # Configuration management
```

## Development

This project uses Go modules for dependency management.

```bash
# Add a new dependency
go get github.com/some/dependency
```

### Adding a New Provider

To add a new AI provider:

1. Create a new file in `pkg/ai/providers/` that implements the `Provider` interface
2. Update the provider factory in `pkg/ai/providers/factory.go`
3. Add provider constants and configuration in `internal/config/config.go`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)

## Support

If you encounter any issues or have questions, please file an issue on the [GitHub repository](https://github.com/dalekurt/kube-ai/issues).

## Security Note

Kube-AI stores API keys in the configuration file. In a production environment, you may want to implement more secure key storage methods or use environment variables for sensitive information. 
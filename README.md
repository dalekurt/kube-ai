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

### Set API Key

To securely store your AI service API key in the system keyring:

```bash
kubectl ai set-key <your-api-key>
```

### Delete API Key

To remove your API key from the system keyring:

```bash
kubectl ai del-key
```

## Configuration

Kube-AI uses the system keyring to securely store your API key. You only need to set it once using the `set-key` command as shown above.

## Project Structure

```
kube-ai/
├── cmd/             # Application entry points
├── pkg/             # Public packages
│   ├── k8s/         # Kubernetes client utilities
│   └── ai/          # AI integration 
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)

## Support

If you encounter any issues or have questions, please file an issue on the [GitHub repository](https://github.com/dalekurt/kube-ai/issues).

## Security Note

Kube-AI uses your system's keyring to store the API key securely. This is generally more secure than storing it in plain text or environment variables. However, the security of the keyring depends on your operating system and its configuration. Always ensure you're following best practices for system security. 
# Release Process for Kube-AI

This document describes the process for creating and publishing a new release of Kube-AI.

## Prerequisites

Before creating a release, ensure you have:

- Git access with permission to push tags
- Task runner installed (`go install github.com/go-task/task/v3/cmd/task@latest`)
- A clean working directory (no uncommitted changes)
- Completed and tested all features for this release
- Updated documentation as needed

## Release Process

### 1. Update CHANGELOG.md (Optional)

You can manually prepare the CHANGELOG.md with detailed descriptions before releasing, but our automated tools will generate entries based on git commits if you don't:

```bash
# Generate or update a changelog entry for a specific version
task changelog:specific -- X.Y.Z

# Automatically calculate the next version and update changelog
task changelog
```

Our tooling:
- Categorizes commits based on conventional commit prefixes
- Groups them into "Added", "Changed", "Fixed", and "Other" sections
- Automatically determines appropriate version bumps based on commit types

### 2. Create the Release

We provide a simple command to create a new release using Task:

```bash
task release -- X.Y.Z
```

Replace `X.Y.Z` with the appropriate semantic version number:
- **X** (Major): Incompatible API changes
- **Y** (Minor): Backward-compatible new features
- **Z** (Patch): Backward-compatible bug fixes

For example:
```bash
task release -- 1.0.0
```

This command will:
1. Validate that the version follows semantic versioning format
2. Check that the version exists in CHANGELOG.md
3. Update the version in the code
4. Create and push a git tag
5. Push all changes to the main branch

### 3. Automated GitHub Actions Release Workflow

Once the tag is pushed, GitHub Actions will automatically:

1. Update CHANGELOG.md with generated entries if they don't exist
2. Build binaries for all supported platforms
3. Package archives with correct formats for distribution
4. Update `ai.yaml` with the new version, download URLs, and SHA256 checksums
5. Commit the updated files back to the repository
6. Extract release notes from CHANGELOG.md
7. Create a GitHub release with all artifacts
8. Add SLSA Level 3 supply chain security metadata

You can monitor the progress at: https://github.com/dalekurt/kube-ai/actions

### 4. Supply Chain Security

Kube-AI implements [SLSA Level 3](https://slsa.dev) supply chain security using GitHub Actions. After the main release workflow completes, the SLSA workflow will:

1. Generate cryptographic provenance information for all artifacts
2. Sign the artifacts with keyless signing
3. Attach the provenance to the GitHub release
4. Provide an attestation that can be verified by users

This ensures that all released binaries are built in a secure, trusted environment with verifiable provenance.

### 5. Docker Image Release

After the GitHub release is complete, you can also publish a Docker image:

```bash
task docker:build
task docker:push
```

This will build and push Docker images with both the version tag and the 'latest' tag.

### 6. Verify the Release

Once the automated processes complete:

1. Verify the GitHub release page contains all expected artifacts
2. Check that the updated CHANGELOG.md and ai.yaml are committed to the repository
3. Verify the SLSA provenance information is attached to the release
4. Download and test the binary for your platform
5. Verify the Docker image works correctly

```bash
# Test the Docker image
docker run --rm dalekurt/kube-ai:X.Y.Z version
```

### 7. Kubernetes Krew Plugin Publishing

After a successful release, the `ai.yaml` file will be automatically updated with the correct version, download URLs and SHA256 checksums. It will be included in the GitHub release assets and committed back to the repository.

To publish to the Krew Plugin Index:
1. Fork the [krew-index](https://github.com/kubernetes-sigs/krew-index) repository
2. Copy the updated `ai.yaml` to `plugins/ai.yaml` in your fork
3. Submit a pull request to the krew-index repository

## Troubleshooting

If you encounter problems during the release process:

### GitHub Actions Failure

1. Check the GitHub Actions logs for specific errors
2. Fix any build or test issues
3. If necessary, delete the tag and re-run the release process:

```bash
git tag -d vX.Y.Z
git push --delete origin vX.Y.Z
```

### Docker Issues

If you encounter Docker build or push issues:

```bash
# Check Docker build
docker build -t dalekurt/kube-ai:test .

# Verify Docker login
docker login
```

## Manual Release Process

If you need to create a release manually:

1. Update version in `pkg/version/version.go`
2. Use the script to generate or extract changelog entry: `./scripts/update-changelog.sh X.Y.Z`
3. Commit changes: `git commit -m "chore: prepare release vX.Y.Z"`
4. Create a tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
5. Push changes and tag: `git push origin main && git push origin vX.Y.Z`
6. Build binaries manually: `task build:all`
7. Create a GitHub release manually, uploading the binaries from the `bin/` directory 
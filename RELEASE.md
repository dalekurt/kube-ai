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

### 1. Update CHANGELOG.md

Ensure that the `CHANGELOG.md` file is up-to-date with all notable changes for this release:

- New features
- Bug fixes
- Performance improvements
- Breaking changes or deprecations

The release script will automatically update the release date, but you should make sure all changes are documented under the "Unreleased" section.

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
1. Update the version in the code
2. Update the CHANGELOG.md with the release date
3. Commit these changes
4. Create and push a git tag
5. Push all changes to the main branch

### 3. GitHub Actions Release Workflow

Once the tag is pushed, GitHub Actions will automatically:

1. Build binaries for all supported platforms
2. Create checksums for verification
3. Generate release notes
4. Create a GitHub release with all artifacts
5. Add supply chain security metadata (SLSA provenance)

You can monitor the progress at: https://github.com/dalekurt/kube-ai/actions

### 4. Docker Image Release

After the GitHub release is complete, you can also publish a Docker image:

```bash
task docker:build
task docker:push
```

This will build and push Docker images with both the version tag and the 'latest' tag.

### 5. Verify the Release

Once the automated processes complete:

1. Verify the GitHub release page contains all expected artifacts
2. Download and test the binary for your platform
3. Verify the Docker image works correctly

```bash
# Test the Docker image
docker run --rm dalekurt/kube-ai:X.Y.Z version
```

### 6. Announce the Release

Consider announcing the new release through appropriate channels:

- Project documentation updates
- Social media
- Relevant community forums

## Post-Release

After a successful release:

1. Update the project roadmap with plans for the next release
2. Close any issues or milestones associated with this release
3. Consider creating issues for any known items to address in the next release

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
2. Update `CHANGELOG.md` with the release date
3. Commit changes: `git commit -m "chore: prepare release vX.Y.Z"`
4. Create a tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
5. Push changes and tag: `git push origin main && git push origin vX.Y.Z`
6. Build binaries manually: `task build:all`
7. Create a GitHub release manually, uploading the binaries from the `bin/` directory 
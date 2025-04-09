# Changelog

All notable changes to the Kube-AI project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.5] - 2025-04-09

### Added
- docs: update README.md with comprehensive feature overview (83b18ba)
- feat: add cmd/kube-ai directory with main.go and commands.go (45a1509)
- chore: add .task/ to .gitignore (51eaa94)
- ci: update GitHub Actions to use Taskfile and add auto-release script (43e7294)
- Improved release format to follow semver standards
- Updated README with comprehensive feature documentation
- Added proper cmd/kube-ai structure
- Fixed GitHub Actions workflow for reliable releases

### Changed
- Fix awk newline issue in update-changelog.sh script (1d2340d)
- Add changelog update tasks to Taskfile.yml (1e8f42f)
- Update README.md to document changelog and release automation scripts (815cbaa)
- chore: prepare release v0.1.4 (cfc9c56)
- chore: prepare release v0.1.3 (26dd041)
- chore: prepare release v0.1.2 (e683b8a)
- chore: prepare release v0.1.1 (8d88240)
- chore: prepare release v0.1.0 (da21dce)
- chore: ignore task checksum directory (a255091)

### Fixed
- fix: update .gitignore to allow cmd/kube-ai directory (c9dce8f)
- fix: update GitHub Actions workflow to build binaries manually (0cdae89)

### Other
- Add script to automate CHANGELOG.md updates based on git commits (b797d0b)
- Add semantic versioning validation to auto-release script (3399a82)
- docs: update release format to match semver standard (d75ed96)
- ci: remove Docker build and push steps from GitHub Actions workflow (022c805)
- docs: update CHANGELOG.md for initial release (607ba84)

## [0.1.4] - 2025-04-09

## [0.1.3] - 2025-04-09

## [0.1.2] - 2025-04-09

## [0.1.1] - 2025-04-09

## [0.1.0] - 2025-04-09

### Added
- Initial release of Kube-AI
- CLI commands for analyzing Kubernetes resources
- Support for various AI providers (OpenAI, Ollama, Anthropic, Gemini, AnythingLLM)
- Log analysis feature for troubleshooting Kubernetes issues
- Resource optimization suggestions
- Scaling strategy recommendations
- Manifest generation from descriptions
- Error explanation capabilities

### Changed
- N/A

### Fixed
- N/A

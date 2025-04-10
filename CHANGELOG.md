# Changelog

All notable changes to the Kube-AI project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.8] - 2025-04-10

### Added
- No new features added in this release

### Changed
- No significant changes in this release

### Fixed
- No fixes in this release

### Other

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

### Added
- No new features added in this release

### Changed
- No significant changes in this release

### Fixed
- No fixes in this release

### Other

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

### Added
- No new features added in this release

### Changed
- No significant changes in this release

### Fixed
- No fixes in this release

### Other

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

### Added
- No new features added in this release

### Changed
- No significant changes in this release

### Fixed
- No fixes in this release

### Other

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

## [0.1.8] - 2025-04-10

## [0.1.7] - 2025-04-19

### Added
- Fix linting errors: handle unchecked errors, remove unused variables, add unused commands to rootCmd (eb6265a)


### Changed
- chore(deps): update softprops/action-gh-release action to v2 (2a0f88c)


### Fixed
- No fixes in this release

### Other
- Updated README.md (682c459)
- Enable SLSA supply chain security by configuring workflow to run on release publication (0a87e1b)
- Add step to commit updated ai.yaml back to the repository during release (56dc9bd)

## [0.1.6] - 2025-04-09

### Added
- No new features added in this release

### Changed
- chore(deps): update actions/setup-go action to v5 (6cac02b)
- chore(deps): update alpine docker tag to v3.21 (8d8fe71)
- chore(deps): update actions/checkout action to v4 (5b10cab)
- chore(deps): update slsa-framework/slsa-github-generator action to v2 (c3ec683)


### Fixed
- No fixes in this release

### Other
- Automate ai.yaml updates in GitHub release workflow (64d7669)

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

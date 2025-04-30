# Curlglish (ncurl) Improvement Plan

## Project Structure
- [x] Create standard Go project layout
  - [x] Implement `cmd/ncurl` for CLI entry point
  - [x] Implement `internal/httpx` for HTTP request/response handling
  - [x] Implement `internal/llm` for Anthropic API wrapper
  - [x] Add `docs` directory for documentation

## Code Quality
- [x] Add appropriate comments and documentation
  - [x] Add package documentation
  - [x] Add godoc examples
- [x] Implement error handling best practices
  - [x] Use custom error types
  - [x] Improve error messages
- [x] Add logging with levels (info, debug, error)
- [x] Implement context propagation for cancellation
- [x] Add configuration management (flags, env vars, config file)

## Testing
- [x] Add unit tests for all packages
  - [x] Add tests for request generation
  - [x] Add tests for HTTP execution
  - [x] Add tests for LLM integration

## CI/CD
- [x] Set up GitHub Actions workflow
  - [x] Implement linting and formatting checks
  - [x] Run tests
  - [x] Check code coverage
  - [x] Add security scanning
- [ ] Set up goreleaser for automated releases
  - [ ] Configure cross-platform builds
  - [ ] Add Homebrew formula
  - [ ] Set up automated changelog
- [x] Add version information to binary

## Documentation
- [x] Improve README with more examples and usage information
- [x] Add CONTRIBUTING.md
- [x] Add CODE_OF_CONDUCT.md
- [x] Add LICENSE file
- [ ] Add CHANGELOG.md
- [ ] Set up GitHub issue templates

## Features
- [x] Add timeout support
- [x] Add verbose mode
- [x] Add output formatting options (JSON, pretty, raw)
- [ ] Implement command history

## Implementation Plan
1. [x] Restructure project to follow standard Go layout
2. [x] Implement basic functionality in new structure
3. [x] Add tests for core functionality
4. [ ] Set up CI/CD
5. [x] Improve documentation
6. [ ] Add additional features

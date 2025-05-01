# ncurl – curl in English ⚡️

[![CI](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml/badge.svg)](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stephenbyrne99/ncurl.svg)](https://pkg.go.dev/github.com/stephenbyrne99/ncurl)

`ncurl` lets you describe an HTTP request in plain language. It asks Anthropic’s Claude to translate the description into a fully-specified request, executes it, and prints a JSON summary with status, headers and body.

---

## 💾 Installation

### Prerequisites

- **Go**: Version 1.22 or higher is required. If you don't have Go installed, follow these instructions:

#### Installing Go

**macOS:**
```bash
# Using Homebrew
brew install go

# Or download the installer from https://go.dev/dl/
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# Fedora
sudo dnf install golang

# Arch Linux
sudo pacman -S go
```

**Windows:**
```bash
# Using Chocolatey
choco install golang

# Or download the installer from https://go.dev/dl/
```

Verify your installation:
```bash
go version
```

### Option 1: Using Go Install

```bash
go install github.com/stephenbyrne99/ncurl/cmd/ncurl@latest
```

### Option 2: Using npm

```bash
npm install -g ncurl
```

This will download the pre-built binary for your platform.

### Option 3: Building from Source

```bash
# Clone the repository
git clone https://github.com/stephenbyrne99/ncurl.git
cd ncurl

# Build using the installation script
./scripts/install.sh
```

The script will build from source and install to `~/.local/bin` (or `~/bin` if `.local/bin` doesn't exist).
If needed, it will also provide instructions to add the installation directory to your PATH.

> **Prerequisite:** You need an Anthropic API key. Export it in your shell before running `ncurl`.

### Setting Up Your API Key

To avoid having to export your API key each time, add it to your shell's configuration file:

**For Bash users (.bashrc):**
```bash
echo 'export ANTHROPIC_API_KEY="your-key-here"' >> ~/.bashrc
source ~/.bashrc
```

**For Zsh users (.zshrc):**
```bash
echo 'export ANTHROPIC_API_KEY="your-key-here"' >> ~/.zshrc
source ~/.zshrc
```

**For Fish users (config.fish):**
```bash
echo 'set -x ANTHROPIC_API_KEY "your-key-here"' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

Replace `"your-key-here"` with your actual Anthropic API key from https://console.anthropic.com/.

---

## 🚀 Quick Start

```bash
# Simple GET
ncurl "download https://httpbin.org/get"

# POST with JSON and a shorter timeout
ncurl -t 10 "POST to httpbin with a name field being hello"

# Pipe prettified JSON through jq
ncurl "get github stephenbyrne99 ncurl repo" | jq '.body | fromjson | .stargazers_count'
```

---

## 🗂️ Project layout

```
ncurl/
├── cmd/
│   └── ncurl/          # CLI entry-point
├── internal/
│   ├── httpx/          # Request struct + executor
│   └── llm/            # Anthropic wrapper
├── go.mod
└── README.md
```

`httpx` and `llm` live in `internal/` so that they remain implementation details—only the top-level command is public.

---

## 🏗️ Continuous Integration

A simple **GitHub Actions** workflow (`.github/workflows/ci.yml`) runs on every push:

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Validate formatting
        run: |
          go vet ./...
          gofmt -s -d $(git ls-files '*.go') | tee /dev/stderr | (! read)
      - name: Run tests (none yet, placeholder)
        run: go test ./...
      - name: Build CLI
        run: go build ./cmd/ncurl
```

Add tests under `internal/...` as the project grows—GitHub will run them automatically.

---

## 📦 Releases

When you’re ready to ship binaries, drop a **GoReleaser** config and add this job to the workflow:

```yaml
  release:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' && startsWith(github.event.head_commit.message, 'release:')
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Each tag that starts with `release:` will build cross-platform binaries, attach them to the GitHub Release page, and update the `go install` path.

---

## 🤝 Contributing

1. Fork & clone, then run `go vet ./...` before opening a PR.
2. Keep commits small and descriptive.
3. All checks must pass before merge.

---

## 📝 License

MIT © 2025 Stephen Byrne

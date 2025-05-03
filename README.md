# ncurl – cURL in natural language 

[![CI](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml/badge.svg)](https://github.com/stephenbyrne99/ncurl/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stephenbyrne99/ncurl.svg)](https://pkg.go.dev/github.com/stephenbyrne99/ncurl)

`ncurl` lets you describe HTTP requests in plain language. It uses Anthropic's Claude to translate your description into a fully-specified request, executes it, and returns the results.

## ✨ Features

- **Natural Language Interface**: Describe API requests however you want
- **Command History**: Save, search, and rerun previous commands
- **Response Handling**: Well-formatted output for both text and binary responses
- **Evaluation Framework**: Test and validate natural language interpretation accuracy
- **JSON Mode**: Output only response bodies for easy piping
- **Verbose Mode**: See the full request details
- **Cross-Platform**: Works on macOS, Linux, and Windows

---

## Quick install

### Option 1: Using Go Install

```bash
go install github.com/stephenbyrne99/ncurl/cmd/ncurl@latest
```

### Option 2: Using npm

```bash
npm install -g @stephen_byrne_/ncurl
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

### Setting Up Your API Key

> **Prerequisite:** You need an Anthropic API key to use ncurl.

Add your API key to your shell's configuration file:

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

**For PowerShell users:**
```powershell
[Environment]::SetEnvironmentVariable("ANTHROPIC_API_KEY", "your-key-here", "User")
```

Replace `"your-key-here"` with your actual Anthropic API key from https://console.anthropic.com/.

---

## 🚀 Quick Start

```bash
# Simple GET
ncurl "download https://httpbin.org/get"

# POST with JSON and a shorter timeout
ncurl -t 10 "POST to httpbin with a name field being hello"

# Use command history
ncurl -history

# Pipe prettified JSON through jq
ncurl "get github stephenbyrne99 ncurl repo" | jq '.body | fromjson | .stargazers_count'
```

## 📋 Command Options

| Option | Description |
|--------|-------------|
| `-t <seconds>` | Set timeout in seconds (default: 30) |
| `-m <model>` | Specify Anthropic model to use (default: claude-3-7-sonnet) |
| `-j` | Output response body as JSON only |
| `-v` | Verbose output (include request details) |
| `-history` | View command history |
| `-search <term>` | Search command history |
| `-rerun <n>` | Rerun the nth command in history |
| `-i` | Interactive history selection |
| `-version` | Show version information |

Check the [usage documentation](docs/usage.md) for more detailed examples.

## 📊 Evaluation Framework

ncurl includes a comprehensive evaluation system to test the accuracy of its natural language interpretation. Use the `ncurl-eval` utility to run test cases and measure performance:

```bash
# Build the evaluation tool
go build -o ncurl-eval ./cmd/ncurl-eval

# Run all built-in test cases
./ncurl-eval

# Save results to a file
./ncurl-eval -output results.md
```

See the [evaluations documentation](docs/evaluations.md) for more information about creating custom test cases and extending the framework.

## 🗂️ Project Structure

```
ncurl/
├── cmd/
│   ├── ncurl/          # CLI entry-point
│   └── ncurl-eval/     # Evaluation tool
├── internal/
│   ├── httpx/          # Request struct + executor
│   ├── llm/            # Anthropic wrapper
│   ├── history/        # Command history management
│   └── evals/          # Evaluation framework
├── docs/               # Documentation
├── go.mod              # Go module definition
└── README.md
```

## 🤝 Contributing

Contributions are welcome! Please check out the [contributing guidelines](CONTRIBUTING.md) for more information.

1. Fork & clone, then run `go vet ./...` before opening a PR.
2. Keep commits small and descriptive.
3. All checks must pass before merge.


## Why?
I wanted to square off codex vs claude code on building something I would find mildly useful, and have them get it to a complete + production level with as little input as possible.

## 📝 License

MIT © 2025 Stephen Byrne

# Installing ncurl

ncurl (natural language curl) is a command-line tool that lets you make HTTP requests using natural language descriptions.

## Prerequisites

- Go 1.22 or later (for building from source)
- An Anthropic API key

## Installation Methods

### Using Homebrew (macOS and Linux)

```bash
brew install stephenbyrne99/ncurl/ncurl
```

### Using NPM (Cross-platform)

```bash
npm install -g nlcurl
```

This will install the `nlcurl` command which is identical to `ncurl`.

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/stephenbyrne99/ncurl.git
   cd ncurl
   ```

2. Build the binary:
   ```bash
   go build -o ncurl ./cmd/ncurl
   ```

3. Move the binary to your PATH:
   ```bash
   sudo mv ncurl /usr/local/bin/
   ```

## Setting Up Your API Key

ncurl requires an Anthropic API key to function. You can get one from [Anthropic's website](https://console.anthropic.com/).

Set your API key as an environment variable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

For persistent use, add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.).
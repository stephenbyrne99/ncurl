# Installing ncurl

ncurl (natural language curl) is a command-line tool that lets you make HTTP requests using natural language descriptions.

## Prerequisites

- Go 1.22 or later (for building from source)
- An Anthropic API key (can be obtained from [Anthropic's website](https://console.anthropic.com/))

## Installation Methods

### Using Homebrew (macOS and Linux)

```bash
brew install stephenbyrne99/ncurl/ncurl
```

After installation, you'll need to set up your API key as described in the [API Key Setup](#api-key-setup) section below.

### Using NPM (Cross-platform)

```bash
npm install -g ncurl
```

### Using the Installation Script (Recommended)

We provide an installation script that handles building, setting up your PATH, and configuring your API key:

```bash
# Clone the repository
git clone https://github.com/stephenbyrne99/ncurl.git
cd ncurl

# Run the installation script
./scripts/install.sh
```

The script will:
1. Build ncurl from source
2. Install it to your local bin directory
3. **Automatically add the bin directory to your PATH**
4. **Help you set up your Anthropic API key**

### Building from Source (Manual Method)

If you prefer to handle the installation manually:

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
   # Option 1: Move to a system directory (requires sudo)
   sudo mv ncurl /usr/local/bin/
   
   # Option 2: Move to a user directory and update PATH
   mv ncurl ~/.local/bin/
   # Then add ~/.local/bin to your PATH if it's not already there
   ```

## <a name="api-key-setup"></a>Setting Up Your API Key

ncurl requires an Anthropic API key to function. You can get one from [Anthropic's website](https://console.anthropic.com/).

**Important:** The tool will not work without this API key.

Set your API key as an environment variable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

For persistent use, add this line to your shell profile:

```bash
# For bash users
echo 'export ANTHROPIC_API_KEY="your-api-key-here"' >> ~/.bashrc
source ~/.bashrc

# For zsh users
echo 'export ANTHROPIC_API_KEY="your-api-key-here"' >> ~/.zshrc
source ~/.zshrc
```

If you used the installation script, it will have offered to set this up for you automatically.

#!/usr/bin/env bash
set -e

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Error: Go is not installed or not in your PATH."
    print_error "Please install Go from https://golang.org/dl/ before continuing."
    exit 1
fi

# Set installation directories
INSTALL_DIR="$HOME/.local/bin"
FALLBACK_DIR="$HOME/bin"

# Check if the installation directory exists, if not use fallback or create it
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -d "$FALLBACK_DIR" ]; then
        print_warning "Directory $INSTALL_DIR does not exist, using $FALLBACK_DIR instead."
        INSTALL_DIR="$FALLBACK_DIR"
    else
        print_warning "Creating directory $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
fi

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

print_success "Building ncurl from source..."

# Navigate to the repository directory
cd "$REPO_DIR"

# Build the binary
go build -o ncurl ./cmd/ncurl
if [ $? -ne 0 ]; then
    print_error "Failed to build ncurl"
    exit 1
fi

print_success "Successfully built ncurl"

# Copy binary to installation directory
cp ncurl "$INSTALL_DIR/ncurl"
chmod +x "$INSTALL_DIR/ncurl"

print_success "Successfully installed ncurl to $INSTALL_DIR/ncurl"

# Check if the directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_warning "The directory $INSTALL_DIR is not in your PATH."
    
    # Detect shell and recommend PATH update command
    SHELL_NAME=$(basename "$SHELL")
    RC_FILE=""
    PATH_EXPORT="export PATH=\"\$PATH:$INSTALL_DIR\""
    
    if [[ "$SHELL_NAME" == "bash" ]]; then
        RC_FILE="$HOME/.bashrc"
    elif [[ "$SHELL_NAME" == "zsh" ]]; then
        RC_FILE="$HOME/.zshrc"
    fi
    
    if [[ -n "$RC_FILE" ]]; then
        print_warning "To add it to your PATH, run:"
        echo "echo '$PATH_EXPORT' >> $RC_FILE"
        echo "source $RC_FILE"
    else
        print_warning "To add it to your PATH, add the following line to your shell configuration file:"
        echo "$PATH_EXPORT"
    fi
    
    print_warning "Or run ncurl directly: $INSTALL_DIR/ncurl"
else
    print_success "The directory is in your PATH. You can run ncurl from anywhere."
fi

print_warning "Note: Remember to set ANTHROPIC_API_KEY in your environment before using ncurl."
print_warning "You can add this to your shell configuration file with:"
if [[ "$SHELL_NAME" == "bash" ]]; then
    echo "echo 'export ANTHROPIC_API_KEY=\"your-api-key\"' >> $HOME/.bashrc"
elif [[ "$SHELL_NAME" == "zsh" ]]; then
    echo "echo 'export ANTHROPIC_API_KEY=\"your-api-key\"' >> $HOME/.zshrc"
else
    echo "export ANTHROPIC_API_KEY=\"your-api-key\""
fi

print_success "Installation complete! Run 'ncurl help' to get started."
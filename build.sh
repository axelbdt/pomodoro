#!/bin/bash
set -e

echo "Building pomodoro timer..."

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install with: sudo apt install golang"
    exit 1
fi

# Tidy dependencies
echo "Downloading dependencies..."
go mod tidy

# Build
echo "Compiling..."
go build -o pomodoro .

echo ""
echo "Build complete: ./pomodoro"
echo ""
echo "To install system-wide:"
echo "  sudo install -m 755 pomodoro /usr/local/bin/"
echo ""
echo "To test:"
echo "  ./pomodoro status"
echo "  ./pomodoro tray"

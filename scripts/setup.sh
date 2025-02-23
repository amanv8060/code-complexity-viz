#!/bin/bash

# Create necessary directories
mkdir -p static templates logs

# Build WASM
GOOS=js GOARCH=wasm go build -o static/analyzer.wasm wasm/analyzer.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/

# Get dependencies
go mod tidy

# Build the project
go build -o code-complexity-viz

echo "Setup complete! Run './code-complexity-viz' to start the server" 
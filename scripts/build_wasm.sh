#!/bin/bash
set -e  # Exit on any error

# Ensure static directory exists
mkdir -p static

# Build WASM binary
echo "Building WASM binary..."
GOOS=js GOARCH=wasm go build -o static/analyzer.wasm wasm/analyzer.go
if [ $? -ne 0 ]; then
    echo "Failed to build WASM binary"
    exit 1
fi

# Copy wasm_exec.js
echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/
if [ $? -ne 0 ]; then
    echo "Failed to copy wasm_exec.js"
    exit 1
fi

echo "WASM build completed successfully"
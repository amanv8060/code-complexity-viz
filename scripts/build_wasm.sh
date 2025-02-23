#!/bin/bash
GOOS=js GOARCH=wasm go build -o static/analyzer.wasm wasm/analyzer.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/ 
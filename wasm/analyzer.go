package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/aman/code-complexity-viz/analyzer"
)

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("analyzeGoCode", js.FuncOf(analyzeGoCode))
	<-c
}

func analyzeGoCode(this js.Value, args []js.Value) (result interface{}) {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			result = wrap("Internal error: "+fmt.Sprint(r), nil)
		}
	}()

	if len(args) < 1 {
		return wrap("Error: No code provided", nil)
	}

	// Validate input type
	if args[0].Type() != js.TypeString {
		return wrap("Error: Input must be a string", nil)
	}

	// Get code from JavaScript
	code := args[0].String()

	// Validate code length
	if len(code) == 0 {
		return wrap("Error: Empty code provided", nil)
	}
	if len(code) > 5000000 { // 5MB limit
		return wrap("Error: Code size exceeds limit", nil)
	}

	// Analyze the code
	fileAnalyzer, err := analyzer.NewFileAnalyzer("temp.go", []byte(code))
	if err != nil {
		return wrap(err.Error(), nil)
	}

	results := fileAnalyzer.AnalyzeFile()
	if len(results) == 0 {
		return wrap("No functions found", nil)
	}

	// Convert results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return wrap(err.Error(), nil)
	}

	return wrap("", string(jsonData))
}

func wrap(err string, data interface{}) js.Value {
	result := make(map[string]interface{})
	if err != "" {
		result["error"] = err
	} else {
		result["data"] = data
	}
	return js.ValueOf(result)
}

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/aman/code-complexity-viz/analyzer"
)

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("analyzeGoCode", js.FuncOf(analyzeGoCode))
	<-c
}

func analyzeGoCode(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return wrap("No code provided", nil)
	}

	// Get code from JavaScript
	code := args[0].String()

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
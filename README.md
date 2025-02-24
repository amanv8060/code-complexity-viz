# Code Complexity Visualizer

A web-based tool for analyzing and visualizing various code complexity metrics in Go source files.

🔗 [Try the Live Demo](https://amanv8060.github.io/code-complexity-viz)

## Usage Modes

### Browser Analysis (WASM)
- Runs entirely in your browser
- No server required
- Perfect for quick analysis
- Available in the live demo

### Server Analysis
- Full feature set
- Handles larger files
- Better performance
- Requires local setup

## Features

### Complexity Metrics
- **Cyclomatic Complexity (McCabe)**: Measures the number of linearly independent paths through code
- **Cognitive Complexity**: Measures how difficult it is to understand the code's control flow
- **Halstead Metrics**:
  - Volume: Measures the size of the implementation
  - Difficulty: Indicates how hard the code is to understand
  - Effort: Estimates the effort required to maintain the code
- **Maintainability Index**: A composite metric indicating overall maintainability (0-100 scale)
- **Lines of Code**: Physical lines of code per function

## Installation

1. Clone the repository:
```bash
git clone https://github.com/aman/code-complexity-viz
cd code-complexity-viz
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run main.go
```

4. Open `http://localhost:8080` in your browser

## Usage

1. Upload a Go source file using the web interface
2. Click "Analyze" to process the file
3. View the visualization of complexity metrics
4. Use the dropdown to switch between different metrics
5. Hover over bars to see detailed metrics for each function

### Limitations
- Maximum file size: 5MB
- Only analyzes `.go` files
- Functions must be syntactically valid Go code


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


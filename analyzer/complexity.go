package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
)

// Operator represents a unique operator in Halstead metrics
type Operator struct {
	Token     token.Token
	Name      string
	Frequency int
}

// Operand represents a unique operand in Halstead metrics
type Operand struct {
	Name      string
	Frequency int
}

// MetricsResult stores the complexity metrics for a single file or function
type MetricsResult struct {
	Name                   string  `json:"name"`
	CyclomaticComplexity  int     `json:"cyclomaticComplexity"`
	CognitiveComplexity   int     `json:"cognitiveComplexity"`
	LinesOfCode           int     `json:"linesOfCode"`
	HalsteadVolume        float64 `json:"halsteadVolume"`
	HalsteadDifficulty    float64 `json:"halsteadDifficulty"`
	HalsteadEffort        float64 `json:"halsteadEffort"`
	MaintainabilityIndex  float64 `json:"maintainabilityIndex"`
	NestingLevel          int     `json:"nestingLevel"`
}

// FileAnalyzer handles the analysis of a single file
type FileAnalyzer struct {
	fset *token.FileSet
	ast  *ast.File
}

// NewFileAnalyzer creates a new analyzer for the given file content
func NewFileAnalyzer(filename string, content []byte) (*FileAnalyzer, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, content, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	return &FileAnalyzer{
		fset: fset,
		ast:  node,
	}, nil
}

// CalculateCyclomaticComplexity calculates McCabe's cyclomatic complexity
func (fa *FileAnalyzer) CalculateCyclomaticComplexity(node ast.Node) int {
	complexity := 1 // Base complexity

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause,
			*ast.CommClause:
			complexity++
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

// CalculateCognitiveComplexity calculates cognitive complexity
func (fa *FileAnalyzer) CalculateCognitiveComplexity(node ast.Node) int {
	complexity := 0
	nestingLevel := 0

	var inspect func(ast.Node) bool
	inspect = func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			if n.Else != nil {
				complexity++
				ast.Inspect(n.Else, inspect)
			}
			nestingLevel--
			return false
		case *ast.ForStmt, *ast.RangeStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n, inspect)
			nestingLevel--
			return false
		case *ast.SwitchStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n, inspect)
			nestingLevel--
			return false
		}
		return true
	}

	ast.Inspect(node, inspect)
	return complexity
}

// calculateHalsteadMetrics calculates Halstead complexity metrics
func (fa *FileAnalyzer) calculateHalsteadMetrics(node ast.Node) (volume, difficulty, effort float64) {
	operators := make(map[token.Token]int)
	operands := make(map[string]int)

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BinaryExpr:
			operators[n.Op]++
		case *ast.UnaryExpr:
			operators[n.Op]++
		case *ast.Ident:
			operands[n.Name]++
		case *ast.BasicLit:
			operands[n.Value]++
		}
		return true
	})

	n1 := float64(len(operators)) // unique operators
	n2 := float64(len(operands))  // unique operands
	N1 := float64(0)             // total operators
	N2 := float64(0)             // total operands

	for _, count := range operators {
		N1 += float64(count)
	}
	for _, count := range operands {
		N2 += float64(count)
	}

	vocabulary := n1 + n2
	length := N1 + N2
	volume = float64(length) * math.Log2(vocabulary)
	difficulty = (n1 / 2) * (N2 / n2)
	effort = difficulty * volume

	return volume, difficulty, effort
}

// CalculateMaintainabilityIndex calculates the maintainability index
func (fa *FileAnalyzer) CalculateMaintainabilityIndex(cyclomatic int, halsteadVolume float64, linesOfCode int) float64 {
	// Original formula: MI = 171 - 5.2 * ln(HV) - 0.23 * CC - 16.2 * ln(LOC)
	mi := 171 - 5.2*math.Log(halsteadVolume) - 0.23*float64(cyclomatic) - 16.2*math.Log(float64(linesOfCode))
	// Normalize to 0-100 scale
	mi = math.Max(0, math.Min(100, mi*100/171))
	return mi
}

// CountLinesOfCode counts the number of non-empty lines in a function
func (fa *FileAnalyzer) CountLinesOfCode(node ast.Node) int {
	start := fa.fset.Position(node.Pos())
	end := fa.fset.Position(node.End())
	return end.Line - start.Line + 1
}

// AnalyzeFunction analyzes a single function and returns its metrics
func (fa *FileAnalyzer) AnalyzeFunction(funcDecl *ast.FuncDecl) *MetricsResult {
	// Add validation for nil function declaration
	if funcDecl == nil || funcDecl.Name == nil {
		return nil
	}

	cyclomaticComplexity := fa.CalculateCyclomaticComplexity(funcDecl)
	cognitiveComplexity := fa.CalculateCognitiveComplexity(funcDecl)
	linesOfCode := fa.CountLinesOfCode(funcDecl)
	
	volume, difficulty, effort := fa.calculateHalsteadMetrics(funcDecl)
	
	// Add validation for edge cases
	if volume <= 0 {
		volume = 1 // Avoid log(0) in maintainability index calculation
	}
	if linesOfCode <= 0 {
		linesOfCode = 1
	}
	
	maintainabilityIndex := fa.CalculateMaintainabilityIndex(cyclomaticComplexity, volume, linesOfCode)

	// Ensure all metrics are valid
	if math.IsNaN(volume) || math.IsInf(volume, 0) {
		volume = 0
	}
	if math.IsNaN(difficulty) || math.IsInf(difficulty, 0) {
		difficulty = 0
	}
	if math.IsNaN(effort) || math.IsInf(effort, 0) {
		effort = 0
	}
	if math.IsNaN(maintainabilityIndex) || math.IsInf(maintainabilityIndex, 0) {
		maintainabilityIndex = 0
	}

	return &MetricsResult{
		Name:                  funcDecl.Name.Name,
		CyclomaticComplexity: cyclomaticComplexity,
		CognitiveComplexity:  cognitiveComplexity,
		LinesOfCode:          linesOfCode,
		HalsteadVolume:       math.Round(volume*100) / 100,
		HalsteadDifficulty:   math.Round(difficulty*100) / 100,
		HalsteadEffort:       math.Round(effort*100) / 100,
		MaintainabilityIndex: math.Round(maintainabilityIndex*100) / 100,
	}
}

// AnalyzeFile analyzes the entire file and returns metrics for all functions
func (fa *FileAnalyzer) AnalyzeFile() []*MetricsResult {
	var results []*MetricsResult

	ast.Inspect(fa.ast, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			results = append(results, fa.AnalyzeFunction(funcDecl))
		}
		return true
	})

	return results
}

package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
)

// FileAnalyzer represents a file analyzer.
type FileAnalyzer struct {
	fset *token.FileSet
	ast  *ast.File
}

func NewFileAnalyzer(filename string, content []byte) (*FileAnalyzer, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &FileAnalyzer{
		fset: fset,
		ast:  node,
	}, nil
}

// AnalyzeFunction analyzes a function declaration and returns the metrics.
func (fa *FileAnalyzer) AnalyzeFunction(funcDecl *ast.FuncDecl) *MetricsResult {
	if funcDecl == nil || funcDecl.Name == nil {
		return nil
	}

	cyclomaticComplexity := fa.CalculateCyclomaticComplexity(funcDecl)
	cognitiveComplexity := fa.CalculateCognitiveComplexity(funcDecl)
	linesOfCode := fa.CountLinesOfCode(funcDecl)
	volume, difficulty, effort := fa.calculateHalsteadMetrics(funcDecl)
	maintainabilityIndex := fa.CalculateMaintainabilityIndex(cyclomaticComplexity, volume, linesOfCode)
	nestedDepth := fa.calculateNestedDepth(funcDecl)
	commentDensity := fa.calculateCommentDensity(funcDecl)
	paramCount := 0
	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			paramCount += len(field.Names)
		}
	}
	returnCount := fa.countReturnStatements(funcDecl)

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
	if math.IsNaN(commentDensity) || math.IsInf(commentDensity, 0) {
		commentDensity = 0
	}

	return &MetricsResult{
		Name:                 funcDecl.Name.Name,
		CyclomaticComplexity: cyclomaticComplexity,
		CognitiveComplexity:  cognitiveComplexity,
		LinesOfCode:          linesOfCode,
		HalsteadVolume:       math.Round(volume*100) / 100,
		HalsteadDifficulty:   math.Round(difficulty*100) / 100,
		HalsteadEffort:       math.Round(effort*100) / 100,
		MaintainabilityIndex: math.Round(maintainabilityIndex*100) / 100,
		NestedDepth:          nestedDepth,
		CommentDensity:       math.Round(commentDensity*100) / 100,
		FunctionParameters:   paramCount,
		ReturnStatements:     returnCount,
	}
}

// AnalyzeFile analyzes a file and returns the metrics for each function.
func (fa *FileAnalyzer) AnalyzeFile() []*MetricsResult {
	var results []*MetricsResult

	ast.Inspect(fa.ast, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			result := fa.AnalyzeFunction(funcDecl)
			if result != nil {
				results = append(results, result)
			}
		}
		return true
	})

	return results
}

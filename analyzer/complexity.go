package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
)

type Operator struct {
	Token     token.Token
	Name      string
	Frequency int
}

// Operand represents an operand in the code.
type Operand struct {
	Name      string
	Frequency int
}

// MetricsResult stores the complexity metrics for a single file or function
type MetricsResult struct {
	Name                 string  `json:"name"`
	CyclomaticComplexity int     `json:"cyclomaticComplexity"` // Cyclomatic complexity of the function.
	CognitiveComplexity  int     `json:"cognitiveComplexity"`  // Cognitive complexity of the function.
	LinesOfCode          int     `json:"linesOfCode"`          // Lines of code in the function.
	HalsteadVolume       float64 `json:"halsteadVolume"`       // Halstead volume of the function.
	HalsteadDifficulty   float64 `json:"halsteadDifficulty"`   // Halstead difficulty of the function.
	HalsteadEffort       float64 `json:"halsteadEffort"`       // Halstead effort of the function.
	MaintainabilityIndex float64 `json:"maintainabilityIndex"` // Maintainability index of the function.
	NestedDepth          int     `json:"nestedDepth"`          // Nested depth of the function.
	CommentDensity       float64 `json:"commentDensity"`       // Comment density of the function.
	FunctionParameters   int     `json:"functionParameters"`   // Number of function parameters.
	ReturnStatements     int     `json:"returnStatements"`     // Number of return statements.
}

// CalculateCyclomaticComplexity calculates the cyclomatic complexity.
func (fa *FileAnalyzer) CalculateCyclomaticComplexity(node ast.Node) int {
	if node == nil {
		return 1
	}

	complexity := 1

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			complexity++
			if n.Else != nil {
				elseStmt := n.Else
				for {
					if elseIf, ok := elseStmt.(*ast.IfStmt); ok {
						complexity++
						elseStmt = elseIf.Else
					} else {
						break
					}
				}
			}
		case *ast.ForStmt, *ast.RangeStmt, *ast.SelectStmt, *ast.TypeSwitchStmt, *ast.SwitchStmt:
			complexity++ // Increment for these control flow constructs

		case *ast.CaseClause:
			// Correctly handle multiple expressions in a single case.
			if n.List != nil {
				complexity += len(n.List) // Increment for *each* expression
			}
		case *ast.CommClause:
			if n.Comm != nil { // not default
				complexity++
			}
		case *ast.FuncDecl: // Ensure that we do not enter other function definition
			if n == node { // we only check node, not its descendant nodes.
				return true
			}
			return false // nested function
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		case *ast.CallExpr:
			// Check if the called function is an operator like "&&" or "||"
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok {
					if id.Name == "builtin" && (sel.Sel.Name == "and" || sel.Sel.Name == "or") { // hypothetical
						complexity++
					}
				}
			} else if id, ok := n.Fun.(*ast.Ident); ok { // for short-circuit evaluation in function call
				if id.Name == "and" || id.Name == "or" { // hypothetical
					complexity++
				}
			}
		}
		return true
	})

	return complexity
}

// CalculateCognitiveComplexity calculates the cognitive complexity of a given node.
func (fa *FileAnalyzer) CalculateCognitiveComplexity(node ast.Node) int {
	if node == nil {
		return 0
	}

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
				if _, isElseIf := n.Else.(*ast.IfStmt); !isElseIf {
					complexity++
				}
				ast.Inspect(n.Else, inspect)
			}
			nestingLevel--
			return false

		case *ast.ForStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			nestingLevel--
			return false
		case *ast.RangeStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			nestingLevel--
			return false
		case *ast.SwitchStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			nestingLevel--
			return false
		case *ast.TypeSwitchStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			nestingLevel--
			return false
		case *ast.SelectStmt:
			complexity += 1 + nestingLevel
			nestingLevel++
			ast.Inspect(n.Body, inspect)
			nestingLevel--
			return false
		case *ast.FuncLit: // closure
			complexity++
		}
		return true
	}

	ast.Inspect(node, inspect)
	return complexity
}

// calculateHalsteadMetrics calculates the Halstead metrics (volume, difficulty, effort) for a given AST node
func (fa *FileAnalyzer) calculateHalsteadMetrics(node ast.Node) (volume, difficulty, effort float64) {
	if node == nil {
		return 0, 0, 0
	}

	temp, err := CalculateHalsteadMetrics(node)
	if err != nil {
		fmt.Println(err)
		return 0, 0, 0
	}

	return temp.Volume, temp.Difficulty, temp.Effort

}

// CalculateMaintainabilityIndex calculates the maintainability index of a given function.
func (fa *FileAnalyzer) CalculateMaintainabilityIndex(cyclomatic int, halsteadVolume float64, linesOfCode int) float64 {
	// formula from https://docs.microsoft.com/en-us/visualstudio/code-quality/maintainability-index
	// Maintainability Index = MAX(0,(171 - 5.2 * ln(Halstead Volume) - 0.23 * (Cyclomatic Complexity) - 16.2 * ln(Lines of Code))*100 / 171)
	if halsteadVolume <= 0 {
		halsteadVolume = 1
	}
	if linesOfCode <= 0 {
		linesOfCode = 1
	}
	mi := 171 - 5.2*math.Log(halsteadVolume) - 0.23*float64(cyclomatic) - 16.2*math.Log(float64(linesOfCode))
	mi = math.Max(0, math.Min(100, math.Round(mi*100/171)))
	return mi
}

// calculateNestedDepth calculates the nested depth of a given node.
func (fa *FileAnalyzer) calculateNestedDepth(node ast.Node) int {
	if node == nil {
		return 0
	}

	maxDepth := 0
	currentDepth := 0

	var inspect func(ast.Node) bool
	inspect = func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		}

		// Continue traversing
		switch n := n.(type) {
		case *ast.IfStmt:
			ast.Inspect(n.Body, inspect)
			if n.Else != nil {
				ast.Inspect(n.Else, inspect)
			}
			currentDepth--
			return false
		case *ast.ForStmt:
			ast.Inspect(n.Body, inspect)
			currentDepth--
			return false
		case *ast.RangeStmt:
			ast.Inspect(n.Body, inspect)
			currentDepth--
			return false
		case *ast.SwitchStmt:
			ast.Inspect(n.Body, inspect)
			currentDepth--
			return false
		case *ast.TypeSwitchStmt:
			ast.Inspect(n.Body, inspect)
			currentDepth--
			return false
		case *ast.SelectStmt:
			ast.Inspect(n.Body, inspect)
			currentDepth--
			return false
		}

		return true
	}

	ast.Inspect(node, inspect)
	return maxDepth
}

// calculateCommentDensity calculates the comment density of a given node.
func (fa *FileAnalyzer) calculateCommentDensity(node ast.Node) float64 {
	if node == nil || fa.ast == nil {
		return 0
	}

	commentCount := 0
	if fa.ast.Comments != nil {
		for _, comment := range fa.ast.Comments {
			commentCount += len(comment.List)
		}
	}
	linesOfCode := fa.CountLinesOfCode(node)
	if linesOfCode <= 0 {
		return 0
	}

	return float64(commentCount) / float64(linesOfCode)
}

// countReturnStatements counts the number of return statements in a given node.
func (fa *FileAnalyzer) countReturnStatements(node ast.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ReturnStmt:
			count++
		case *ast.FuncLit:
			// we don't want to count return statements inside a closure
			return false // Stop traversing this node.
		}
		return true // Continue traversing other nodes.
	})

	return count
}

// CountLinesOfCode counts the lines of code in a given node.
func (fa *FileAnalyzer) CountLinesOfCode(node ast.Node) int {
	if node == nil {
		return 1
	}
	start := fa.fset.Position(node.Pos())
	end := fa.fset.Position(node.End())

	// Ensure that end line is never less than start line
	if end.Line < start.Line {
		return 1 // Return a minimum of 1 line
	}

	var docStringComments int
	if nd := node.(*ast.FuncDecl); nd != nil {
		if nd.Doc != nil {
			startDoc := fa.fset.Position(nd.Doc.Pos())
			endDoc := fa.fset.Position(nd.Doc.End())
			if endDoc.Line < startDoc.Line {
				docStringComments = 1
			}
			docStringComments = endDoc.Line - startDoc.Line + 1
		}

	}

	return end.Line - start.Line + 1 + docStringComments
}

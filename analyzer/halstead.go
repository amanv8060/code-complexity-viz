package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
)

// HalsteadMetrics (struct - no changes)
type HalsteadMetrics struct {
	N1         int // Total number of operators
	N2         int // Total number of operands
	Eta1       int // Number of distinct operators
	Eta2       int // Number of distinct operands
	Length     int
	Vocabulary int
	Volume     float64
	Difficulty float64
	Effort     float64
}

// CalculateHalsteadMetrics (updated)
func CalculateHalsteadMetrics(node ast.Node) (HalsteadMetrics, error) {
	if node == nil {
		return HalsteadMetrics{}, fmt.Errorf("input node cannot be nil")
	}

	operators := make(map[token.Token]int)
	operands := make(map[string]int)

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		// Operators
		case *ast.BinaryExpr:
			operators[x.Op]++
		case *ast.UnaryExpr:
			operators[x.Op]++
		case *ast.CallExpr:
			operators[token.FUNC]++
			if x.Ellipsis.IsValid() {
				operators[token.ELLIPSIS]++
			}
		case *ast.IncDecStmt:
			operators[x.Tok]++
		case *ast.AssignStmt:
			operators[x.Tok]++
		case *ast.ReturnStmt:
			operators[token.RETURN]++
		case *ast.IfStmt:
			operators[token.IF]++
		case *ast.ForStmt:
			operators[token.FOR]++
		case *ast.RangeStmt:
			operators[token.RANGE]++
		case *ast.SwitchStmt:
			operators[token.SWITCH]++
		case *ast.CaseClause:
			operators[token.CASE]++
		case *ast.TypeSwitchStmt:
			operators[token.SWITCH]++
		case *ast.TypeAssertExpr:
			operators[token.PERIOD]++
		case *ast.SendStmt:
			operators[token.ARROW]++
		case *ast.GoStmt:
			operators[token.GO]++
		case *ast.DeferStmt:
			operators[token.DEFER]++
		case *ast.BranchStmt:
			operators[x.Tok]++
		case *ast.SelectorExpr:
			operators[token.PERIOD]++
		//Parenthesis are not counted as operators in many implementations
		//Operands
		case *ast.Ident:
			if x.Obj != nil && x.Obj.Kind == ast.Typ {
				operands[x.Name]++
			} else {
				operands[x.Name]++
			}
		case *ast.BasicLit:
			operands[x.Value]++
		case *ast.CompositeLit:
			if x.Type != nil {
				if _, ok := x.Type.(*ast.Ellipsis); ok {
					operators[token.ELLIPSIS]++
				} else if id, ok := x.Type.(*ast.Ident); ok {
					operands[id.Name]++
				}

			}
		case *ast.FuncDecl:
			//Function name as an operand
			operands[x.Name.Name]++

		case *ast.ChanType:
			operands["chan"]++

		case *ast.FuncType:
			if x.Params != nil {
				for _, field := range x.Params.List {
					if _, ok := field.Type.(*ast.Ellipsis); ok {
						operators[token.ELLIPSIS]++
					}
				}
			}
		}
		return true
	})

	N1 := 0
	for _, count := range operators {
		N1 += count
	}

	N2 := 0
	for _, count := range operands {
		N2 += count
	}

	Eta1 := len(operators)
	Eta2 := len(operands)

	length := N1 + N2
	vocabulary := Eta1 + Eta2

	// Halstead calculations (with adjustments)
	var volume, difficulty, effort float64

	if vocabulary > 0 {
		volume = float64(length) * math.Log2(float64(vocabulary))
	}
	if Eta2 > 0 {
		difficulty = (float64(Eta1) / 2.0) * (float64(N2) / float64(Eta2))
	}
	effort = difficulty * volume

	metrics := HalsteadMetrics{
		N1:         N1,
		N2:         N2,
		Eta1:       Eta1,
		Eta2:       Eta2,
		Length:     length,
		Vocabulary: vocabulary,
		Volume:     volume,
		Difficulty: difficulty,
		Effort:     effort,
	}

	return metrics, nil
}

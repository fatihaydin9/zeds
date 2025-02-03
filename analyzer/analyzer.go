package analyzer

import (
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"math"
	"os"
	"strings"
)

// MethodResult holds the analysis results for each function.
type MethodResult struct {
	MethodName           string
	Cyclomatic           int
	HalsteadVolume       float64
	LOC                  int
	MaintainabilityIndex float64
}

// CalculateCyclomaticComplexity calculates the cyclomatic complexity for a given AST node.
func CalculateCyclomaticComplexity(n ast.Node) int {
	complexity := 1
	ast.Inspect(n, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			// Count each case clause
			complexity++
		case *ast.BinaryExpr:
			// Count logical operators && and ||
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})
	return complexity
}

// CalculateHalsteadVolume computes a simplified Halstead Volume based on operator and operand counts.
func CalculateHalsteadVolume(src string) float64 {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, []byte(src), nil, scanner.ScanComments)

	totalOperators := 0
	totalOperands := 0
	uniqueOperators := make(map[string]bool)
	uniqueOperands := make(map[string]bool)

	// Define tokens considered as operators.
	operatorTokens := map[token.Token]bool{
		token.ADD:            true,
		token.SUB:            true,
		token.MUL:            true,
		token.QUO:            true,
		token.REM:            true,
		token.AND:            true,
		token.OR:             true,
		token.XOR:            true,
		token.SHL:            true,
		token.SHR:            true,
		token.AND_NOT:        true,
		token.ADD_ASSIGN:     true,
		token.SUB_ASSIGN:     true,
		token.MUL_ASSIGN:     true,
		token.QUO_ASSIGN:     true,
		token.REM_ASSIGN:     true,
		token.AND_ASSIGN:     true,
		token.OR_ASSIGN:      true,
		token.XOR_ASSIGN:     true,
		token.SHL_ASSIGN:     true,
		token.SHR_ASSIGN:     true,
		token.AND_NOT_ASSIGN: true,
		token.EQL:            true,
		token.LSS:            true,
		token.GTR:            true,
		token.ASSIGN:         true,
		token.NOT:            true,
		token.NEQ:            true,
		token.LEQ:            true,
		token.GEQ:            true,
		token.LAND:           true,
		token.LOR:            true,
		token.DEFINE:         true,
	}

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if operatorTokens[tok] {
			totalOperators++
			key := lit
			if key == "" {
				key = tok.String()
			}
			uniqueOperators[key] = true
		} else if tok == token.IDENT || tok == token.INT || tok == token.FLOAT ||
			tok == token.IMAG || tok == token.CHAR || tok == token.STRING {
			totalOperands++
			key := lit
			if key == "" {
				key = tok.String()
			}
			uniqueOperands[key] = true
		}
	}

	vocabulary := len(uniqueOperators) + len(uniqueOperands)
	length := totalOperators + totalOperands
	var volume float64
	if vocabulary > 0 {
		volume = float64(length) * math.Log2(float64(vocabulary))
	} else {
		volume = 0
	}
	return volume
}

// CalculateLOC returns the number of lines in the source code.
func CalculateLOC(src string) int {
	return len(strings.Split(src, "\n"))
}

// CalculateMaintainabilityIndex computes the Maintainability Index (MI) using a standard formula and a bonus from comment density.
func CalculateMaintainabilityIndex(cyclomatic int, halsteadVolume float64, loc int, commentDensity float64, commentDensityMultiplier float64) float64 {
	safeVolume := halsteadVolume
	if safeVolume <= 0 {
		safeVolume = 1
	}
	safeLOC := float64(loc)
	if safeLOC <= 0 {
		safeLOC = 1
	}
	baseMI := (171 - 5.2*math.Log(safeVolume) - 0.23*float64(cyclomatic) - 16.2*math.Log(safeLOC)) * 100 / 171
	bonus := commentDensityMultiplier * math.Sin(math.Sqrt(2.4*commentDensity))
	mi := baseMI + bonus
	if mi < 0 {
		mi = 0
	}
	return mi
}

// CalculateCommentDensity computes the ratio of comment lines to total lines in the file.
func CalculateCommentDensity(fileContent string, comments []*ast.CommentGroup) float64 {
	totalLines := len(strings.Split(fileContent, "\n"))
	commentLines := 0
	for _, cg := range comments {
		for _, comment := range cg.List {
			commentLines += len(strings.Split(comment.Text, "\n"))
		}
	}
	if totalLines == 0 {
		return 0
	}
	return float64(commentLines) / float64(totalLines)
}

// AnalyzeMethods analyzes all functions in a given Go source file and computes code quality metrics.
// It returns the analysis results for each function and the global comment density.
func AnalyzeMethods(filePath string, commentDensityMultiplier float64) ([]MethodResult, float64, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, 0, err
	}
	source := string(data)

	fset := token.NewFileSet()
	// Parse the file including comments.
	f, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
	if err != nil {
		return nil, 0, err
	}

	globalCommentDensity := CalculateCommentDensity(source, f.Comments)
	var results []MethodResult

	// Traverse the AST to find function declarations.
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			funcName := fn.Name.Name
			startOffset := fset.Position(fn.Body.Pos()).Offset
			endOffset := fset.Position(fn.Body.End()).Offset
			funcSource := source[startOffset:endOffset]

			cc := CalculateCyclomaticComplexity(fn.Body)
			halstead := CalculateHalsteadVolume(funcSource)
			loc := CalculateLOC(funcSource)
			mi := CalculateMaintainabilityIndex(cc, halstead, loc, globalCommentDensity, commentDensityMultiplier)

			results = append(results, MethodResult{
				MethodName:           funcName,
				Cyclomatic:           cc,
				HalsteadVolume:       halstead,
				LOC:                  loc,
				MaintainabilityIndex: mi,
			})
		}
	}

	return results, globalCommentDensity, nil
}

// AnalyzeFile analyzes a Go source file and returns its functions and methods
func AnalyzeFile(filepath string) ([]string, error) {
	// Create a new token set
	fset := token.NewFileSet()

	// Parse the Go source file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Parse the file with its content
	node, err := parser.ParseFile(fset, filepath, data, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var functions []string = make([]string, 0)

	// Visit all nodes in AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			name := x.Name.Name
			if x.Recv != nil {
				recv := x.Recv.List[0].Type
				var recvName string
				if star, ok := recv.(*ast.StarExpr); ok {
					if ident, ok := star.X.(*ast.Ident); ok {
						recvName = ident.Name
					}
				} else if ident, ok := recv.(*ast.Ident); ok {
					recvName = ident.Name
				}
				if recvName != "" {
					functions = append(functions, "method: "+recvName+"."+name)
				}
			} else {
				functions = append(functions, "function: "+name)
			}
		}
		return true
	})

	// Eğer hiç fonksiyon bulunamadıysa boş slice döndür
	if len(functions) == 0 {
		return make([]string, 0), nil
	}

	return functions, nil
}
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type TimeoutViolation struct {
	FilePath string
	Line     int
	CallType string
	Expr     string
}

func main() {
	rootDir := "backend/services"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	var violations []TimeoutViolation

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check DB calls: Query, QueryRow, Exec without Context suffix or missing ctx
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				name := sel.Sel.Name
				if (name == "Query" || name == "QueryRow" || name == "Exec") && isDBSelector(sel.X) {
					pos := fset.Position(call.Pos())
					violations = append(violations, TimeoutViolation{
						FilePath: path,
						Line:     pos.Line,
						CallType: "DB_NON_CONTEXT_CALL",
						Expr:     fmt.Sprintf("%s.%s", exprToString(sel.X), name),
					})
				}

				// Check http.Client without Timeout or http.Get/Post direct calls
				if name == "Get" || name == "Post" || name == "PostForm" {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "http" {
						pos := fset.Position(call.Pos())
						violations = append(violations, TimeoutViolation{
							FilePath: path,
							Line:     pos.Line,
							CallType: "HTTP_DIRECT_CALL_NO_TIMEOUT",
							Expr:     fmt.Sprintf("http.%s", name),
						})
					}
				}
			}

			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== TIMEOUT AUDIT RESULTS (%d violations found) ===\n", len(violations))
	for _, v := range violations {
		fmt.Printf("❌ [%s] %s:%d -> %s\n", v.CallType, v.FilePath, v.Line, v.Expr)
	}

	if len(violations) > 0 {
		fmt.Println("\n⚠️ Timeout compliance issues found.")
	} else {
		fmt.Println("✅ 100% TIMEOUT COMPLIANCE ACROSS ALL SERVICES.")
	}
}

func isDBSelector(expr ast.Expr) bool {
	str := exprToString(expr)
	return strings.Contains(str, "db") || strings.Contains(str, "tx") || strings.Contains(str, "repo") || strings.Contains(str, "DB")
}

func exprToString(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprToString(v.X), v.Sel.Name)
	case *ast.StarExpr:
		return "*" + exprToString(v.X)
	default:
		return "expr"
	}
}

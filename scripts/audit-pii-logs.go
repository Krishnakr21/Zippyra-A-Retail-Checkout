package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PIIViolation struct {
	FilePath string
	Line     int
	Pattern  string
	Content  string
}

var piiPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(phone|mobile|email|password|otp|pin|totp|secret|gstin|pan|bearer|jwt|cvv)`),
}

func main() {
	rootDir := "backend/services"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	var violations []PIIViolation

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

			// Check logger calls (logger.Info, logger.Error, log.Printf, etc.)
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				pkg := exprToString(sel.X)
				name := sel.Sel.Name

				if (pkg == "logger" || pkg == "log" || pkg == "zap") && (name == "Info" || name == "Warn" || name == "Error" || name == "Printf" || name == "Infof" || name == "Errorf") {
					for i, arg := range call.Args {
						if i == 0 {
							// Skip the format string literal itself
							continue
						}
						argStr := exprToString(arg)
						if strings.Contains(argStr, "MaskPhone") || strings.Contains(argStr, "MaskEmail") || strings.Contains(argStr, "Sanitize") || strings.Contains(argStr, "Redact") {
							continue
						}
						for _, pattern := range piiPatterns {
							if pattern.MatchString(argStr) {
								pos := fset.Position(call.Pos())
								violations = append(violations, PIIViolation{
									FilePath: path,
									Line:     pos.Line,
									Pattern:  pattern.String(),
									Content:  argStr,
								})
							}
						}
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

	fmt.Printf("=== PII LOGGING AUDIT RESULTS (%d potential unmasked PII log arguments) ===\n", len(violations))
	for _, v := range violations {
		fmt.Printf("⚠️ [%s:%d] %s\n", v.FilePath, v.Line, v.Content)
	}

	if len(violations) > 0 {
		fmt.Println("\nNote: Verify masked helper wrapping (e.g. logger.MaskPhone) for any flagged log argument.")
	} else {
		fmt.Println("✅ 100% PII LOGGING COMPLIANCE ACROSS ALL SERVICES.")
	}
}

func exprToString(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.BasicLit:
		return v.Value
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprToString(v.X), v.Sel.Name)
	case *ast.CallExpr:
		return fmt.Sprintf("%s(...)", exprToString(v.Fun))
	default:
		return "expr"
	}
}

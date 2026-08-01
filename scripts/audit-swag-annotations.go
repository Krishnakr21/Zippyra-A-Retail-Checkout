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

type MissingSwagViolation struct {
	FilePath string
	Line     int
	Func     string
}

func main() {
	rootDir := "backend/services"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	var violations []MissingSwagViolation
	var annotatedCount int

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}

			funcName := fn.Name.Name
			// Check HTTP handler naming pattern (e.g. Handle*, *Handler)
			if strings.HasPrefix(funcName, "Handle") || strings.HasSuffix(funcName, "Handler") {
				if funcName == "Handle" || funcName == "NewHandler" || funcName == "RegisterRoutes" || funcName == "SetupRoutes" {
					continue
				}

				hasRouter := false
				hasSummary := false

				if fn.Doc != nil {
					for _, comment := range fn.Doc.List {
						text := comment.Text
						if strings.Contains(text, "@Router") {
							hasRouter = true
						}
						if strings.Contains(text, "@Summary") {
							hasSummary = true
						}
					}
				}

				pos := fset.Position(fn.Pos())
				if hasRouter && hasSummary {
					annotatedCount++
				} else {
					violations = append(violations, MissingSwagViolation{
						FilePath: path,
						Line:     pos.Line,
						Func:     funcName,
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== SWAG ANNOTATION COVERAGE AUDIT ===\n")
	fmt.Printf("Annotated Handlers: %d\n", annotatedCount)
	fmt.Printf("Unannotated Handlers: %d\n", len(violations))

	if len(violations) > 0 {
		fmt.Println("\n⚠️ Missing Swag Annotations on the following HTTP Handlers:")
		for i, v := range violations {
			if i < 20 { // Cap output for readability
				fmt.Printf("   %s:%d -> %s()\n", v.FilePath, v.Line, v.Func)
			}
		}
		if len(violations) > 20 {
			fmt.Printf("   ... and %d more handlers\n", len(violations)-20)
		}
	} else {
		fmt.Println("✅ 100% SWAG ANNOTATION COVERAGE ACROSS ALL HTTP HANDLERS.")
	}
}

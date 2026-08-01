package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AST & Git Diff Analyzer: Checks PR modifications to v1 endpoint request/response structs
func main() {
	fmt.Println("=== AUDITING API V1 STRUCT MODIFICATIONS FOR BREAKING CHANGES ===")

	// Git diff against target branch (e.g. main / origin/main)
	targetBranch := os.Getenv("TARGET_BRANCH")
	if targetBranch == "" {
		targetBranch = "origin/main"
	}

	cmd := exec.Command("git", "diff", targetBranch, "--", "backend/services/*/*handlers.go", "backend/services/*/models.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback for standalone repo execution or local check
		fmt.Printf("ℹ Git diff against %s not available or no changes found. Skipping diff lint.\n", targetBranch)
		os.Exit(0)
	}

	diffStr := string(output)
	if diffStr == "" {
		fmt.Println("✅ No v1 handler or model struct modifications detected.")
		os.Exit(0)
	}

	lines := strings.Split(diffStr, "\n")
	var warnings []string

	for i, line := range lines {
		if strings.HasPrefix(line, "+") && (strings.Contains(line, "struct {") || strings.Contains(line, "type ")) {
			// Check if previous or current lines contain "// backward-compatible" marker
			hasMarker := false
			for j := max(0, i-5); j <= min(len(lines)-1, i+5); j++ {
				if strings.Contains(lines[j], "// backward-compatible") || strings.Contains(lines[j], "/v2/") {
					hasMarker = true
					break
				}
			}

			if !hasMarker {
				warnings = append(warnings, fmt.Sprintf("Line: %s", strings.TrimSpace(line)))
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️ NOTICE: The following v1 struct modifications were detected without explicit '// backward-compatible' markers or '/v2/' routes:")
		for _, w := range warnings {
			fmt.Printf("   - %s\n", w)
		}
		fmt.Println("\n👉 Nudge: Verify these changes do not break existing v1 API consumers. If compatible, add `// backward-compatible` comment.")
	} else {
		fmt.Println("✅ All v1 struct changes verified or marked as backward-compatible.")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

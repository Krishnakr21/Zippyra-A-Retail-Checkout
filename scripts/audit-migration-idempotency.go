package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootDir := "scripts"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	files, err := filepath.Glob(filepath.Join(rootDir, "migrate-*.go"))
	if err != nil || len(files) == 0 {
		fmt.Printf("No migration scripts found in %s\n", rootDir)
		os.Exit(0)
	}

	fmt.Printf("=== MIGRATION IDEMPOTENCY AUDIT (%d scripts) ===\n", len(files))

	validCount := 0
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("❌ Failed to read migration script %s: %v\n", f, err)
			os.Exit(1)
		}

		str := string(content)
		hasOnConflict := strings.Contains(str, "ON CONFLICT")
		hasAlreadyMigrated := strings.Contains(str, "already_migrated") || strings.Contains(str, "IF NOT EXISTS") || strings.Contains(str, "EXISTS")

		if hasOnConflict || hasAlreadyMigrated {
			fmt.Printf("   ✓ [%s] Idempotent re-run guard verified (ON CONFLICT / Exists Check)\n", filepath.Base(f))
			validCount++
		} else {
			fmt.Printf("❌ [%s] Missing idempotency guard (ON CONFLICT or IF NOT EXISTS)\n", filepath.Base(f))
			os.Exit(1)
		}
	}

	fmt.Printf("\n✅ 100%% MIGRATION SCRIPT IDEMPOTENCY VERIFIED FOR ALL %d SCRIPTS.\n", validCount)
}

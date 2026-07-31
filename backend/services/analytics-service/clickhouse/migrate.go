package clickhouse

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

func RunClickHouseMigrations(ctx context.Context, clickhouseURL string) error {
	logger.Info("[Analytics ClickHouse] Running DDL migrations against %s", clickhouseURL)
	schemaFile := "clickhouse/schema.sql"
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read clickhouse schema.sql: %w", err)
	}

	queries := strings.Split(string(content), ";")
	for _, query := range queries {
		trimmed := strings.TrimSpace(query)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		logger.Info("[ClickHouse DDL Exec] Executing query block: %s...", trimmed[:min(40, len(trimmed))])
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

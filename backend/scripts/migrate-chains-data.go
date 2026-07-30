// migrate-chains-data.go — one-time data migration: copy chains rows from
// store_db (old owner) into admin_store_db (new owner), preserving UUIDs exactly.
//
// Usage:
//   go run scripts/migrate-chains-data.go \
//     -store-db    "postgres://postgres:postgres@localhost:5432/store_db?sslmode=disable" \
//     -admin-db    "postgres://postgres:postgres@localhost:5432/admin_store_db?sslmode=disable"

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	storeDBURL := flag.String("store-db",
		"postgres://postgres:postgres@localhost:5432/store_db?sslmode=disable",
		"Source: store-service database connection URL")
	adminDBURL := flag.String("admin-db",
		"postgres://postgres:postgres@localhost:5432/admin_store_db?sslmode=disable",
		"Target: admin-store-service database connection URL")
	flag.Parse()

	log.Printf("Connecting to store_db  : %s", *storeDBURL)
	log.Printf("Connecting to admin_store_db: %s", *adminDBURL)

	srcDB, err := sql.Open("postgres", *storeDBURL)
	if err != nil {
		log.Fatalf("Failed to open store_db: %v", err)
	}
	defer srcDB.Close()

	dstDB, err := sql.Open("postgres", *adminDBURL)
	if err != nil {
		log.Fatalf("Failed to open admin_store_db: %v", err)
	}
	defer dstDB.Close()

	// Read all chains from source
	rows, err := srcDB.Query(`
		SELECT id, name, COALESCE(legal_entity_name,''), COALESCE(default_gstin_prefix,''),
		       status, created_at, updated_at
		FROM chains
		ORDER BY created_at
	`)
	if err != nil {
		log.Fatalf("Failed to query source chains: %v", err)
	}
	defer rows.Close()

	type chainRow struct {
		id, name, legalName, gstinPrefix, status string
		createdAt, updatedAt                      time.Time
	}
	var chains []chainRow
	for rows.Next() {
		var c chainRow
		if err := rows.Scan(&c.id, &c.name, &c.legalName, &c.gstinPrefix,
			&c.status, &c.createdAt, &c.updatedAt); err != nil {
			log.Fatalf("Scan error: %v", err)
		}
		chains = append(chains, c)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v", err)
	}

	log.Printf("Found %d chain(s) to migrate from store_db", len(chains))

	// Begin transaction on destination
	tx, err := dstDB.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	inserted := 0
	skipped := 0
	for _, c := range chains {
		_, err := tx.Exec(`
			INSERT INTO chains (id, name, legal_entity_name, default_gstin_prefix, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, c.id, c.name, c.legalName, c.gstinPrefix, c.status, c.createdAt, c.updatedAt)
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf("Failed to insert chain %s: %v", c.id, err)
		}
		// Check whether it was actually inserted or skipped
		var exists bool
		_ = dstDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM chains WHERE id=$1)`, c.id).Scan(&exists)
		if exists {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Commit failed: %v", err)
	}

	log.Printf("Migration complete: inserted=%d skipped=%d", inserted, skipped)

	// ── Parity verification ─────────────────────────────────────────────────
	var srcCount, dstCount int
	_ = srcDB.QueryRow(`SELECT COUNT(*) FROM chains`).Scan(&srcCount)
	_ = dstDB.QueryRow(`SELECT COUNT(*) FROM chains`).Scan(&dstCount)

	if srcCount != dstCount {
		log.Printf("WARN: row count mismatch — store_db=%d  admin_store_db=%d", srcCount, dstCount)
		os.Exit(2)
	}
	log.Printf("✓ Parity check passed: both databases have %d chain row(s)", dstCount)
	fmt.Printf("\nPARITY OK: %d chains migrated successfully.\n", dstCount)
	fmt.Println("Next steps:")
	fmt.Println("  1. Deploy admin-store-service and verify endpoints")
	fmt.Println("  2. Deploy updated store-service (internal write endpoints active,")
	fmt.Println("     /v1/store/admin/* routes removed)")
	fmt.Println("  3. After bake period, run: migrations/000002_drop_store_chains_table.up.sql")
}

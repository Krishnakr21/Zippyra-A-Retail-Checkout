package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	ErrChainNotFound  = errors.New("chain not found")
	ErrChainSuspended = errors.New("chain suspended")
)

// ChainRepository is the interface for chains table operations — admin-store-service owns this domain.
type ChainRepository interface {
	CreateChain(ctx context.Context, c *Chain) error
	GetChainByID(ctx context.Context, id string) (*Chain, error)
	ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error)
	UpdateChain(ctx context.Context, c *Chain) error
	UpdateChainStatus(ctx context.Context, id, status string) error
}

// ── PostgresChainRepository ───────────────────────────────────────────────────

type PostgresChainRepository struct {
	db *sql.DB
}

func NewPostgresChainRepository(db *sql.DB) ChainRepository {
	return &PostgresChainRepository{db: db}
}

func (p *PostgresChainRepository) CreateChain(ctx context.Context, c *Chain) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = ChainStatusActive
	}

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO chains (id, name, legal_entity_name, default_gstin_prefix, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.ID, c.Name, c.LegalEntityName, c.DefaultGstInPrefix, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (p *PostgresChainRepository) GetChainByID(ctx context.Context, id string) (*Chain, error) {
	var c Chain
	err := p.db.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(legal_entity_name,''), COALESCE(default_gstin_prefix,''), status, created_at, updated_at
		 FROM chains WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.LegalEntityName, &c.DefaultGstInPrefix, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *PostgresChainRepository) ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "1=1"
	args := []interface{}{}
	if statusFilter != "" {
		where += " AND status = $1"
		args = append(args, statusFilter)
	}

	var total int64
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM chains WHERE %s", where)
	if err := p.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, pageSize, offset)

	dataQ := fmt.Sprintf(
		`SELECT id, name, COALESCE(legal_entity_name,''), COALESCE(default_gstin_prefix,''), status, created_at, updated_at
		 FROM chains WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, limitIdx, offsetIdx,
	)
	rows, err := p.db.QueryContext(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var chains []*Chain
	for rows.Next() {
		var c Chain
		if err := rows.Scan(&c.ID, &c.Name, &c.LegalEntityName, &c.DefaultGstInPrefix, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		chains = append(chains, &c)
	}
	if chains == nil {
		chains = []*Chain{}
	}
	return chains, total, nil
}

func (p *PostgresChainRepository) UpdateChain(ctx context.Context, c *Chain) error {
	now := time.Now()
	_, err := p.db.ExecContext(ctx,
		`UPDATE chains
		 SET name = COALESCE(NULLIF($1,''), name),
		     legal_entity_name = COALESCE(NULLIF($2,''), legal_entity_name),
		     default_gstin_prefix = COALESCE(NULLIF($3,''), default_gstin_prefix),
		     status = COALESCE(NULLIF($4,''), status),
		     updated_at = $5
		 WHERE id = $6`,
		c.Name, c.LegalEntityName, c.DefaultGstInPrefix, c.Status, now, c.ID,
	)
	return err
}

func (p *PostgresChainRepository) UpdateChainStatus(ctx context.Context, id, status string) error {
	now := time.Now()
	result, err := p.db.ExecContext(ctx,
		`UPDATE chains SET status = $1, updated_at = $2 WHERE id = $3`,
		status, now, id,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrChainNotFound
	}
	return nil
}

// ── MemoryChainRepository for unit tests ─────────────────────────────────────

type MemoryChainRepository struct {
	mu     sync.RWMutex
	chains map[string]*Chain
}

func NewMemoryChainRepository() ChainRepository {
	return &MemoryChainRepository{
		chains: make(map[string]*Chain),
	}
}

func (m *MemoryChainRepository) CreateChain(ctx context.Context, c *Chain) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = ChainStatusActive
	}
	m.chains[c.ID] = c
	return nil
}

func (m *MemoryChainRepository) GetChainByID(ctx context.Context, id string) (*Chain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.chains[id]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (m *MemoryChainRepository) ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Chain
	for _, c := range m.chains {
		if statusFilter != "" && c.Status != statusFilter {
			continue
		}
		cp := *c
		result = append(result, &cp)
	}
	total := int64(len(result))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= len(result) {
		return []*Chain{}, total, nil
	}
	end := offset + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *MemoryChainRepository) UpdateChain(ctx context.Context, c *Chain) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.chains[c.ID]
	if !ok {
		return ErrChainNotFound
	}
	if c.Name != "" {
		existing.Name = c.Name
	}
	if c.LegalEntityName != "" {
		existing.LegalEntityName = c.LegalEntityName
	}
	if c.DefaultGstInPrefix != "" {
		existing.DefaultGstInPrefix = c.DefaultGstInPrefix
	}
	if c.Status != "" {
		existing.Status = c.Status
	}
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryChainRepository) UpdateChainStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.chains[id]
	if !ok {
		return ErrChainNotFound
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}

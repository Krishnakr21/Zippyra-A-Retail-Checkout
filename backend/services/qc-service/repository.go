package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Repository interface {
	CreateReview(ctx context.Context, review *QCReview) error
	GetReviewByGRNID(ctx context.Context, grnID string) (*QCReview, error)
	UpdateReviewLineItems(ctx context.Context, grnID string, updates []QCLineItemUpdate, reviewerID string) (*QCReview, error)
	IsReviewComplete(ctx context.Context, grnID string) (bool, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateReview(ctx context.Context, review *QCReview) error {
	if review.ID == "" {
		review.ID = uuid.New().String()
	}

	lineItemsJSON, err := json.Marshal(review.LineItems)
	if err != nil {
		return fmt.Errorf("failed to marshal line_items: %w", err)
	}

	if review.OverallStatus == "" {
		review.OverallStatus = OverallStatusPending
	}
	now := time.Now()
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}
	review.UpdatedAt = now

	// ON CONFLICT (grn_id) DO NOTHING for idempotency
	query := `
		INSERT INTO qc_reviews (id, grn_id, store_id, line_items, overall_status, reviewed_by, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (grn_id) DO NOTHING
	`

	_, err = r.db.ExecContext(
		ctx, query,
		review.ID, review.GRNID, review.StoreID, lineItemsJSON,
		review.OverallStatus, review.ReviewedBy, review.CompletedAt,
		review.CreatedAt, review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert qc_review: %w", err)
	}

	return nil
}

func (r *postgresRepo) GetReviewByGRNID(ctx context.Context, grnID string) (*QCReview, error) {
	query := `
		SELECT id, grn_id, store_id, line_items, overall_status, reviewed_by, completed_at, created_at, updated_at
		FROM qc_reviews
		WHERE grn_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, grnID)

	var rev QCReview
	var lineItemsBytes []byte

	err := row.Scan(
		&rev.ID, &rev.GRNID, &rev.StoreID, &lineItemsBytes,
		&rev.OverallStatus, &rev.ReviewedBy, &rev.CompletedAt,
		&rev.CreatedAt, &rev.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query qc_review: %w", err)
	}

	if err := json.Unmarshal(lineItemsBytes, &rev.LineItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal line_items json: %w", err)
	}

	return &rev, nil
}

func (r *postgresRepo) UpdateReviewLineItems(ctx context.Context, grnID string, updates []QCLineItemUpdate, reviewerID string) (*QCReview, error) {
	rev, err := r.GetReviewByGRNID(ctx, grnID)
	if err != nil {
		return nil, err
	}
	if rev == nil {
		return nil, fmt.Errorf("review not found for grn_id %s", grnID)
	}

	updateMap := make(map[string]QCLineItemUpdate)
	for _, u := range updates {
		updateMap[u.GRNLineItemID] = u
	}

	allComplete := true
	for i := range rev.LineItems {
		item := &rev.LineItems[i]
		if u, found := updateMap[item.GRNLineItemID]; found {
			item.QCStatus = u.QCStatus
			if u.QCNote != nil {
				item.QCNote = u.QCNote
			}
		}
		if item.QCStatus == QCStatusPending {
			allComplete = false
		}
	}

	now := time.Now()
	rev.UpdatedAt = now
	if reviewerID != "" {
		rev.ReviewedBy = &reviewerID
	}

	if allComplete {
		rev.OverallStatus = OverallStatusComplete
		if rev.CompletedAt == nil {
			rev.CompletedAt = &now
		}
	} else {
		rev.OverallStatus = OverallStatusPending
	}

	lineItemsJSON, err := json.Marshal(rev.LineItems)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated line_items: %w", err)
	}

	updateQuery := `
		UPDATE qc_reviews
		SET line_items = $1, overall_status = $2, reviewed_by = $3, completed_at = $4, updated_at = $5
		WHERE grn_id = $6
	`
	_, err = r.db.ExecContext(ctx, updateQuery, lineItemsJSON, rev.OverallStatus, rev.ReviewedBy, rev.CompletedAt, rev.UpdatedAt, grnID)
	if err != nil {
		return nil, fmt.Errorf("failed to update qc_review: %w", err)
	}

	return rev, nil
}

func (r *postgresRepo) IsReviewComplete(ctx context.Context, grnID string) (bool, error) {
	query := `
		SELECT overall_status
		FROM qc_reviews
		WHERE grn_id = $1
	`
	var status string
	err := r.db.QueryRowContext(ctx, query, grnID).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == OverallStatusComplete, nil
}

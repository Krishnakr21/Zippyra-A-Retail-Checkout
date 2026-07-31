package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrTicketAlreadyClosed = errors.New("TICKET_ALREADY_CLOSED")
	ErrReopenWindowExpired = errors.New("REOPEN_WINDOW_EXPIRED")
	ErrResolutionRequired  = errors.New("RESOLUTION_NOTE_REQUIRED")
	ErrOrderNotOwned       = errors.New("ORDER_NOT_OWNED_BY_REQUESTER")
)

type SupportRepository interface {
	CreateTicket(ctx context.Context, ticket *SupportTicket) error
	GetTicketByID(ctx context.Context, id string) (*SupportTicket, error)
	ListTicketsByRequester(ctx context.Context, requesterID string, statusFilter string, page, limit int) ([]*SupportTicket, error)
	ListTicketsForAgent(ctx context.Context, statusFilter, priorityFilter, agentIDFilter, storeIDFilter string, page, limit int) ([]*SupportTicket, error)
	ListOverdueTickets(ctx context.Context) ([]*SupportTicket, error)
	ListTicketsNearSLA(ctx context.Context) ([]*SupportTicket, error)
	MarkSLAWarned(ctx context.Context, ticketID string) error
	UpdateTicket(ctx context.Context, ticket *SupportTicket) error

	AddMessage(ctx context.Context, msg *TicketMessage) error
	ListMessages(ctx context.Context, ticketID string, includeInternal bool) ([]*TicketMessage, error)
	GetAutoPriority(ctx context.Context, category Category) (Priority, error)

	CreateFeedback(ctx context.Context, fb *FeedbackSubmission) error
	ListFeedback(ctx context.Context, sourceApp string, minScore *int, page, limit int) ([]*FeedbackSubmission, int, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTicket(ctx context.Context, t *SupportTicket) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	query := `
		INSERT INTO support_tickets (id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, false, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.RequesterID, t.RequesterType, t.StoreID, t.ChainID, t.Category, t.RelatedOrderID, t.Subject, t.Description, t.Priority, t.Status, t.AssignedAgentID, t.ResolvedAt, t.ClosedAt, t.SLADueAt)
	return err
}

func (r *PostgresRepository) GetTicketByID(ctx context.Context, id string) (*SupportTicket, error) {
	query := `SELECT id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at
	          FROM support_tickets WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	t := &SupportTicket{}
	if err := row.Scan(&t.ID, &t.RequesterID, &t.RequesterType, &t.StoreID, &t.ChainID, &t.Category, &t.RelatedOrderID, &t.Subject, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID, &t.ResolvedAt, &t.ClosedAt, &t.SLADueAt, &t.IsSLAWarned, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *PostgresRepository) ListTicketsByRequester(ctx context.Context, requesterID string, statusFilter string, page, limit int) ([]*SupportTicket, error) {
	offset := (page - 1) * limit
	query := `SELECT id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at
	          FROM support_tickets WHERE requester_id = $1`
	args := []interface{}{requesterID}
	if statusFilter != "" {
		query += ` AND status = $2`
		args = append(args, statusFilter)
		query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*SupportTicket
	for rows.Next() {
		t := &SupportTicket{}
		if err := rows.Scan(&t.ID, &t.RequesterID, &t.RequesterType, &t.StoreID, &t.ChainID, &t.Category, &t.RelatedOrderID, &t.Subject, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID, &t.ResolvedAt, &t.ClosedAt, &t.SLADueAt, &t.IsSLAWarned, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepository) ListTicketsForAgent(ctx context.Context, statusFilter, priorityFilter, agentIDFilter, storeIDFilter string, page, limit int) ([]*SupportTicket, error) {
	offset := (page - 1) * limit
	query := `SELECT id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at
	          FROM support_tickets WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if statusFilter != "" {
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, statusFilter)
		argIdx++
	}
	if priorityFilter != "" {
		query += fmt.Sprintf(` AND priority = $%d`, argIdx)
		args = append(args, priorityFilter)
		argIdx++
	}
	if agentIDFilter != "" {
		query += fmt.Sprintf(` AND assigned_agent_id = $%d`, argIdx)
		args = append(args, agentIDFilter)
		argIdx++
	}
	if storeIDFilter != "" {
		query += fmt.Sprintf(` AND store_id = $%d`, argIdx)
		args = append(args, storeIDFilter)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY sla_due_at ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*SupportTicket
	for rows.Next() {
		t := &SupportTicket{}
		if err := rows.Scan(&t.ID, &t.RequesterID, &t.RequesterType, &t.StoreID, &t.ChainID, &t.Category, &t.RelatedOrderID, &t.Subject, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID, &t.ResolvedAt, &t.ClosedAt, &t.SLADueAt, &t.IsSLAWarned, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepository) ListOverdueTickets(ctx context.Context) ([]*SupportTicket, error) {
	query := `SELECT id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at
	          FROM support_tickets WHERE status NOT IN ('RESOLVED', 'CLOSED') AND sla_due_at < NOW() ORDER BY sla_due_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*SupportTicket
	for rows.Next() {
		t := &SupportTicket{}
		if err := rows.Scan(&t.ID, &t.RequesterID, &t.RequesterType, &t.StoreID, &t.ChainID, &t.Category, &t.RelatedOrderID, &t.Subject, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID, &t.ResolvedAt, &t.ClosedAt, &t.SLADueAt, &t.IsSLAWarned, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepository) ListTicketsNearSLA(ctx context.Context) ([]*SupportTicket, error) {
	// Find tickets crossing 80% SLA window with is_sla_warned = false
	query := `
		SELECT id, requester_id, requester_type, store_id, chain_id, category, related_order_id, subject, description, priority, status, assigned_agent_id, resolved_at, closed_at, sla_due_at, is_sla_warned, created_at, updated_at
		FROM support_tickets
		WHERE status NOT IN ('RESOLVED', 'CLOSED')
		  AND is_sla_warned = false
		  AND NOW() >= (created_at + 0.8 * (sla_due_at - created_at));
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*SupportTicket
	for rows.Next() {
		t := &SupportTicket{}
		if err := rows.Scan(&t.ID, &t.RequesterID, &t.RequesterType, &t.StoreID, &t.ChainID, &t.Category, &t.RelatedOrderID, &t.Subject, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID, &t.ResolvedAt, &t.ClosedAt, &t.SLADueAt, &t.IsSLAWarned, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepository) MarkSLAWarned(ctx context.Context, ticketID string) error {
	query := `UPDATE support_tickets SET is_sla_warned = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, ticketID)
	return err
}

func (r *PostgresRepository) UpdateTicket(ctx context.Context, t *SupportTicket) error {
	query := `
		UPDATE support_tickets
		SET priority = $1, status = $2, assigned_agent_id = $3, resolved_at = $4, closed_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query, t.Priority, t.Status, t.AssignedAgentID, t.ResolvedAt, t.ClosedAt, t.ID)
	return err
}

func (r *PostgresRepository) AddMessage(ctx context.Context, msg *TicketMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	attachmentsJSON, _ := json.Marshal(msg.Attachments)
	query := `
		INSERT INTO ticket_messages (id, ticket_id, sender_id, sender_type, body, is_internal_note, attachments, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.TicketID, msg.SenderID, msg.SenderType, msg.Body, msg.IsInternalNote, attachmentsJSON)
	return err
}

func (r *PostgresRepository) ListMessages(ctx context.Context, ticketID string, includeInternal bool) ([]*TicketMessage, error) {
	query := `SELECT id, ticket_id, sender_id, sender_type, body, is_internal_note, attachments, created_at
	          FROM ticket_messages WHERE ticket_id = $1`
	if !includeInternal {
		query += ` AND is_internal_note = false`
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*TicketMessage
	for rows.Next() {
		m := &TicketMessage{}
		var attachmentsJSON []byte
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.SenderType, &m.Body, &m.IsInternalNote, &attachmentsJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(attachmentsJSON, &m.Attachments)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *PostgresRepository) GetAutoPriority(ctx context.Context, category Category) (Priority, error) {
	query := `SELECT default_priority FROM ticket_auto_priority_rules WHERE category = $1`
	var p Priority
	err := r.db.QueryRowContext(ctx, query, category).Scan(&p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PriorityNormal, nil
		}
		return PriorityNormal, err
	}
	return p, nil
}

// -------------------------------------------------------------------
// In-Memory Repository Fallback (for unit tests / mock mode)
// -------------------------------------------------------------------
type MemoryRepository struct {
	tickets       map[string]*SupportTicket
	messages      map[string][]*TicketMessage
	autoPriority map[Category]Priority
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tickets:  make(map[string]*SupportTicket),
		messages: make(map[string][]*TicketMessage),
		autoPriority: map[Category]Priority{
			CategoryExitGateIssue: PriorityUrgent,
			CategoryPaymentIssue:  PriorityHigh,
			CategoryOrderIssue:    PriorityNormal,
			CategoryAccountIssue:  PriorityNormal,
			CategoryAppBug:        PriorityLow,
			CategoryDeviceIssue:   PriorityNormal,
			CategoryOther:         PriorityLow,
		},
	}
}

func (r *MemoryRepository) CreateTicket(ctx context.Context, t *SupportTicket) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()
	r.tickets[t.ID] = t
	return nil
}

func (r *MemoryRepository) GetTicketByID(ctx context.Context, id string) (*SupportTicket, error) {
	if t, ok := r.tickets[id]; ok {
		return t, nil
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) ListTicketsByRequester(ctx context.Context, requesterID string, statusFilter string, page, limit int) ([]*SupportTicket, error) {
	var list []*SupportTicket
	for _, t := range r.tickets {
		if t.RequesterID == requesterID {
			if statusFilter == "" || string(t.Status) == statusFilter {
				list = append(list, t)
			}
		}
	}
	return list, nil
}

func (r *MemoryRepository) ListTicketsForAgent(ctx context.Context, statusFilter, priorityFilter, agentIDFilter, storeIDFilter string, page, limit int) ([]*SupportTicket, error) {
	var list []*SupportTicket
	for _, t := range r.tickets {
		if statusFilter != "" && string(t.Status) != statusFilter {
			continue
		}
		if priorityFilter != "" && string(t.Priority) != priorityFilter {
			continue
		}
		if agentIDFilter != "" && (t.AssignedAgentID == nil || *t.AssignedAgentID != agentIDFilter) {
			continue
		}
		if storeIDFilter != "" && (t.StoreID == nil || *t.StoreID != storeIDFilter) {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *MemoryRepository) ListOverdueTickets(ctx context.Context) ([]*SupportTicket, error) {
	var list []*SupportTicket
	now := time.Now()
	for _, t := range r.tickets {
		if t.Status != StatusResolved && t.Status != StatusClosed && now.After(t.SLADueAt) {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *MemoryRepository) ListTicketsNearSLA(ctx context.Context) ([]*SupportTicket, error) {
	var list []*SupportTicket
	now := time.Now()
	for _, t := range r.tickets {
		if t.Status != StatusResolved && t.Status != StatusClosed && !t.IsSLAWarned {
			totalWindow := t.SLADueAt.Sub(t.CreatedAt)
			elapsed := now.Sub(t.CreatedAt)
			if float64(elapsed)/float64(totalWindow) >= 0.8 {
				list = append(list, t)
			}
		}
	}
	return list, nil
}

func (r *MemoryRepository) MarkSLAWarned(ctx context.Context, ticketID string) error {
	if t, ok := r.tickets[ticketID]; ok {
		t.IsSLAWarned = true
	}
	return nil
}

func (r *MemoryRepository) UpdateTicket(ctx context.Context, t *SupportTicket) error {
	t.UpdatedAt = time.Now()
	r.tickets[t.ID] = t
	return nil
}

func (r *MemoryRepository) AddMessage(ctx context.Context, msg *TicketMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreatedAt = time.Now()
	r.messages[msg.TicketID] = append(r.messages[msg.TicketID], msg)
	return nil
}

func (r *MemoryRepository) ListMessages(ctx context.Context, ticketID string, includeInternal bool) ([]*TicketMessage, error) {
	var list []*TicketMessage
	for _, m := range r.messages[ticketID] {
		if !includeInternal && m.IsInternalNote {
			continue
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *MemoryRepository) GetAutoPriority(ctx context.Context, category Category) (Priority, error) {
	if p, ok := r.autoPriority[category]; ok {
		return p, nil
	}
	return PriorityNormal, nil
}

func (r *PostgresRepository) isSQLite() bool {
	return r.db != nil && fmt.Sprintf("%T", r.db.Driver()) == "*sqlite3.SQLiteDriver"
}

func (r *PostgresRepository) CreateFeedback(ctx context.Context, fb *FeedbackSubmission) error {
	if fb.ID == "" {
		fb.ID = uuid.New().String()
	}
	if fb.CreatedAt.IsZero() {
		fb.CreatedAt = time.Now()
	}
	if fb.Context == "" {
		fb.Context = "general"
	}
	query := `INSERT INTO feedback_submissions (id, user_id, user_type, source_app, nps_score, comment, context, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if r.isSQLite() {
		query = `INSERT INTO feedback_submissions (id, user_id, user_type, source_app, nps_score, comment, context, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	}
	_, err := r.db.ExecContext(ctx, query, fb.ID, fb.UserID, fb.UserType, fb.SourceApp, fb.NPSScore, fb.Comment, fb.Context, fb.CreatedAt)
	return err
}

func (r *PostgresRepository) ListFeedback(ctx context.Context, sourceApp string, minScore *int, page, limit int) ([]*FeedbackSubmission, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var whereClauses []string
	var args []interface{}
	argID := 1

	if sourceApp != "" {
		placeholder := fmt.Sprintf("$%d", argID)
		if r.isSQLite() {
			placeholder = "?"
		}
		whereClauses = append(whereClauses, fmt.Sprintf("source_app = %s", placeholder))
		args = append(args, sourceApp)
		argID++
	}
	if minScore != nil {
		placeholder := fmt.Sprintf("$%d", argID)
		if r.isSQLite() {
			placeholder = "?"
		}
		whereClauses = append(whereClauses, fmt.Sprintf("nps_score >= %s", placeholder))
		args = append(args, *minScore)
		argID++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM feedback_submissions%s", whereSQL)
	var total int
	_ = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	limitPlaceholder := fmt.Sprintf("$%d", argID)
	offsetPlaceholder := fmt.Sprintf("$%d", argID+1)
	if r.isSQLite() {
		limitPlaceholder = "?"
		offsetPlaceholder = "?"
	}

	query := fmt.Sprintf(`SELECT id, user_id, user_type, source_app, nps_score, comment, context, created_at
						  FROM feedback_submissions%s ORDER BY created_at DESC LIMIT %s OFFSET %s`, whereSQL, limitPlaceholder, offsetPlaceholder)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*FeedbackSubmission
	for rows.Next() {
		var f FeedbackSubmission
		if err := rows.Scan(&f.ID, &f.UserID, &f.UserType, &f.SourceApp, &f.NPSScore, &f.Comment, &f.Context, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &f)
	}
	return list, total, nil
}

func (r *MemoryRepository) CreateFeedback(ctx context.Context, fb *FeedbackSubmission) error {
	if fb.ID == "" {
		fb.ID = uuid.New().String()
	}
	if fb.CreatedAt.IsZero() {
		fb.CreatedAt = time.Now()
	}
	if fb.Context == "" {
		fb.Context = "general"
	}
	r.tickets[fb.ID] = nil // dummy marker if needed
	return nil
}

func (r *MemoryRepository) ListFeedback(ctx context.Context, sourceApp string, minScore *int, page, limit int) ([]*FeedbackSubmission, int, error) {
	return []*FeedbackSubmission{}, 0, nil
}

var _ = pq.Array

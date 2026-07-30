package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrIRNRecordNotFound = errors.New("IRN record not found")
	ErrDPDPNotFound      = errors.New("DPDP request not found")
	ErrKYCNotFound       = errors.New("KYC record not found")
)

type Repository interface {
	// IRN
	CreateIRNRecord(ctx context.Context, rec *IRNRecord) (bool, error)
	GetIRNRecordByOrderID(ctx context.Context, orderID string) (*IRNRecord, error)
	ListIRNRecords(ctx context.Context, storeID, status string) ([]*IRNRecord, error)
	UpdateIRNStatus(ctx context.Context, id, status string, irn, ackNo *string, ackDate *time.Time, signedQR, response, failureReason *string) error
	IncrementIRNRetry(ctx context.Context, id string, failureReason string) error
	GetFailedIRNRecordsForRetry(ctx context.Context, maxRetries int) ([]*IRNRecord, error)

	// IRN Outbox
	InsertIRNOutbox(ctx context.Context, topic string, payload []byte) error
	GetPendingOutbox(ctx context.Context, limit int) ([]*IRNOutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, id string) error

	// DPDP Consents
	UpsertConsent(ctx context.Context, consent *DPDPConsent) error
	GetLatestConsentsByUser(ctx context.Context, userID string) ([]*DPDPConsent, error)

	// DPDP Requests & Audit
	CreateDPDPRequest(ctx context.Context, req *DPDPRequest) error
	GetDPDPRequestByID(ctx context.Context, id string) (*DPDPRequest, error)
	ListDPDPRequests(ctx context.Context, userID, status, reqType string) ([]*DPDPRequest, error)
	UpdateDPDPRequestStatus(ctx context.Context, id, status string, handledBy, reason *string) error
	InsertDeletionAudit(ctx context.Context, audit *DPDPDeletionAudit) error
	GetDeletionAuditsByRequestID(ctx context.Context, reqID string) ([]*DPDPDeletionAudit, error)

	// DPDP Access Exports
	CreateAccessExport(ctx context.Context, export *DPDPAccessExport) error
	GetAccessExportByRequestID(ctx context.Context, requestID string) (*DPDPAccessExport, error)
	UpdateAccessExportSections(ctx context.Context, export *DPDPAccessExport) error
	UpdateAccessExportStatus(ctx context.Context, export *DPDPAccessExport) error
	SweepExpiredAccessExports(ctx context.Context) (int, error)

	// Merchant KYC
	GetMerchantKYC(ctx context.Context, storeID string) (*MerchantKYC, error)
	UpsertMerchantKYC(ctx context.Context, kyc *MerchantKYC) error
	ListIncompleteKYC(ctx context.Context) ([]*MerchantKYC, error)

	// Velocity Alerts
	CreateVelocityAlert(ctx context.Context, alert *VelocityAlert) error
	ListVelocityAlerts(ctx context.Context, storeID string, unresolvedOnly bool) ([]*VelocityAlert, error)
	ResolveVelocityAlert(ctx context.Context, id string) error

	// Settlement Reconciliation Reports
	SaveSettlementReport(ctx context.Context, report *SettlementReport) error
	GetSettlementReportByDate(ctx context.Context, dateStr string) (*SettlementReport, error)
	ListSettlementReports(ctx context.Context, dateFrom, dateTo string) ([]*SettlementReport, error)
}

// MemoryRepository for unit testing
type MemoryRepository struct {
	mu             sync.RWMutex
	irnRecords     map[string]*IRNRecord
	outboxEvents   []*IRNOutboxEvent
	consents       []*DPDPConsent
	dpdpRequests   map[string]*DPDPRequest
	accessExports  map[string]*DPDPAccessExport
	deletionAudits []*DPDPDeletionAudit
	merchantKYCs   map[string]*MerchantKYC
	velocityAlerts map[string]*VelocityAlert
	reports        map[string]*SettlementReport
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		irnRecords:     make(map[string]*IRNRecord),
		dpdpRequests:   make(map[string]*DPDPRequest),
		accessExports:  make(map[string]*DPDPAccessExport),
		merchantKYCs:   make(map[string]*MerchantKYC),
		velocityAlerts: make(map[string]*VelocityAlert),
		reports:        make(map[string]*SettlementReport),
	}
}

func (m *MemoryRepository) CreateIRNRecord(ctx context.Context, rec *IRNRecord) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.irnRecords {
		if existing.OrderID == rec.OrderID {
			return false, nil // ON CONFLICT DO NOTHING
		}
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	rec.CreatedAt = time.Now().UTC()
	m.irnRecords[rec.ID] = rec
	return true, nil
}

func (m *MemoryRepository) GetIRNRecordByOrderID(ctx context.Context, orderID string) (*IRNRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, rec := range m.irnRecords {
		if rec.OrderID == orderID {
			return rec, nil
		}
	}
	return nil, ErrIRNRecordNotFound
}

func (m *MemoryRepository) ListIRNRecords(ctx context.Context, storeID, status string) ([]*IRNRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*IRNRecord
	for _, rec := range m.irnRecords {
		if storeID != "" && rec.StoreID != storeID {
			continue
		}
		if status != "" && rec.Status != status {
			continue
		}
		res = append(res, rec)
	}
	return res, nil
}

func (m *MemoryRepository) UpdateIRNStatus(ctx context.Context, id, status string, irn, ackNo *string, ackDate *time.Time, signedQR, response, failureReason *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.irnRecords[id]
	if !ok {
		return ErrIRNRecordNotFound
	}
	rec.Status = status
	if irn != nil {
		rec.IRN = irn
	}
	if ackNo != nil {
		rec.AckNo = ackNo
	}
	if ackDate != nil {
		rec.AckDate = ackDate
	}
	if signedQR != nil {
		rec.SignedQRCode = signedQR
	}
	if response != nil {
		rec.IRPResponse = response
	}
	if failureReason != nil {
		rec.FailureReason = failureReason
	}
	now := time.Now().UTC()
	if status == IRNStatusIssued {
		rec.IssuedAt = &now
	}
	return nil
}

func (m *MemoryRepository) IncrementIRNRetry(ctx context.Context, id string, failureReason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.irnRecords[id]
	if !ok {
		return ErrIRNRecordNotFound
	}
	rec.RetryCount++
	rec.FailureReason = &failureReason
	rec.Status = IRNStatusFailed
	return nil
}

func (m *MemoryRepository) GetFailedIRNRecordsForRetry(ctx context.Context, maxRetries int) ([]*IRNRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*IRNRecord
	for _, rec := range m.irnRecords {
		if rec.Status == IRNStatusFailed && rec.RetryCount < maxRetries {
			res = append(res, rec)
		}
	}
	return res, nil
}

func (m *MemoryRepository) InsertIRNOutbox(ctx context.Context, topic string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event := &IRNOutboxEvent{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   string(payload),
		CreatedAt: time.Now().UTC(),
	}
	m.outboxEvents = append(m.outboxEvents, event)
	return nil
}

func (m *MemoryRepository) GetPendingOutbox(ctx context.Context, limit int) ([]*IRNOutboxEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*IRNOutboxEvent
	for _, ev := range m.outboxEvents {
		if ev.PublishedAt == nil {
			res = append(res, ev)
			if len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func (m *MemoryRepository) MarkOutboxPublished(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ev := range m.outboxEvents {
		if ev.ID == id {
			now := time.Now().UTC()
			ev.PublishedAt = &now
			return nil
		}
	}
	return nil
}

func (m *MemoryRepository) UpsertConsent(ctx context.Context, consent *DPDPConsent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if consent.ID == "" {
		consent.ID = uuid.New().String()
	}
	consent.CreatedAt = time.Now().UTC()
	m.consents = append(m.consents, consent)
	return nil
}

func (m *MemoryRepository) GetLatestConsentsByUser(ctx context.Context, userID string) ([]*DPDPConsent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	latestMap := make(map[string]*DPDPConsent)
	for _, c := range m.consents {
		if c.UserID == userID {
			existing, ok := latestMap[c.ConsentType]
			if !ok || c.CreatedAt.After(existing.CreatedAt) {
				latestMap[c.ConsentType] = c
			}
		}
	}

	var res []*DPDPConsent
	for _, c := range latestMap {
		res = append(res, c)
	}
	return res, nil
}

func (m *MemoryRepository) CreateDPDPRequest(ctx context.Context, req *DPDPRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now().UTC()
	m.dpdpRequests[req.ID] = req
	return nil
}

func (m *MemoryRepository) GetDPDPRequestByID(ctx context.Context, id string) (*DPDPRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	req, ok := m.dpdpRequests[id]
	if !ok {
		return nil, ErrDPDPNotFound
	}
	return req, nil
}

func (m *MemoryRepository) ListDPDPRequests(ctx context.Context, userID, status, reqType string) ([]*DPDPRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*DPDPRequest
	for _, req := range m.dpdpRequests {
		if userID != "" && req.UserID != userID {
			continue
		}
		if status != "" && req.Status != status {
			continue
		}
		if reqType != "" && req.RequestType != reqType {
			continue
		}
		res = append(res, req)
	}
	return res, nil
}

func (m *MemoryRepository) UpdateDPDPRequestStatus(ctx context.Context, id, status string, handledBy, reason *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.dpdpRequests[id]
	if !ok {
		return ErrDPDPNotFound
	}
	req.Status = status
	if handledBy != nil {
		req.HandledBy = handledBy
	}
	if reason != nil {
		req.RejectionReason = reason
	}
	if status == DPDPStatusCompleted {
		now := time.Now().UTC()
		req.CompletedAt = &now
	}
	return nil
}

func (m *MemoryRepository) InsertDeletionAudit(ctx context.Context, audit *DPDPDeletionAudit) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if audit.ID == "" {
		audit.ID = uuid.New().String()
	}
	audit.ExecutedAt = time.Now().UTC()
	m.deletionAudits = append(m.deletionAudits, audit)
	return nil
}

func (m *MemoryRepository) GetDeletionAuditsByRequestID(ctx context.Context, reqID string) ([]*DPDPDeletionAudit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*DPDPDeletionAudit
	for _, a := range m.deletionAudits {
		if a.DPDPRequestID == reqID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *MemoryRepository) GetMerchantKYC(ctx context.Context, storeID string) (*MerchantKYC, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	kyc, ok := m.merchantKYCs[storeID]
	if !ok {
		return nil, ErrKYCNotFound
	}
	return kyc, nil
}

func (m *MemoryRepository) UpsertMerchantKYC(ctx context.Context, kyc *MerchantKYC) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	kyc.UpdatedAt = time.Now().UTC()
	m.merchantKYCs[kyc.StoreID] = kyc
	return nil
}

func (m *MemoryRepository) ListIncompleteKYC(ctx context.Context) ([]*MerchantKYC, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*MerchantKYC
	for _, kyc := range m.merchantKYCs {
		if kyc.KYCStatus != "VERIFIED" {
			res = append(res, kyc)
		}
	}
	return res, nil
}

func (m *MemoryRepository) CreateVelocityAlert(ctx context.Context, alert *VelocityAlert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now().UTC()
	m.velocityAlerts[alert.ID] = alert
	return nil
}

func (m *MemoryRepository) ListVelocityAlerts(ctx context.Context, storeID string, unresolvedOnly bool) ([]*VelocityAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*VelocityAlert
	for _, alert := range m.velocityAlerts {
		if storeID != "" && alert.StoreID != storeID {
			continue
		}
		if unresolvedOnly && alert.ResolvedAt != nil {
			continue
		}
		res = append(res, alert)
	}
	return res, nil
}

func (m *MemoryRepository) ResolveVelocityAlert(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, ok := m.velocityAlerts[id]
	if !ok {
		return errors.New("velocity alert not found")
	}
	now := time.Now().UTC()
	alert.ResolvedAt = &now
	return nil
}

func (m *MemoryRepository) SaveSettlementReport(ctx context.Context, report *SettlementReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if report.ID == "" {
		report.ID = uuid.New().String()
	}
	report.GeneratedAt = time.Now().UTC()
	m.reports[report.ReportDate] = report
	return nil
}

func (m *MemoryRepository) GetSettlementReportByDate(ctx context.Context, dateStr string) (*SettlementReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report, ok := m.reports[dateStr]
	if !ok {
		return nil, errors.New("reconciliation report not found for date")
	}
	return report, nil
}

func (m *MemoryRepository) ListSettlementReports(ctx context.Context, dateFrom, dateTo string) ([]*SettlementReport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*SettlementReport
	for _, r := range m.reports {
		if dateFrom != "" && r.ReportDate < dateFrom {
			continue
		}
		if dateTo != "" && r.ReportDate > dateTo {
			continue
		}
		res = append(res, r)
	}
	return res, nil
}

// DPDP Access Export Memory Repository Implementations
func (m *MemoryRepository) CreateAccessExport(ctx context.Context, export *DPDPAccessExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if export.ID == "" {
		export.ID = uuid.New().String()
	}
	export.CreatedAt = time.Now().UTC()
	export.UpdatedAt = time.Now().UTC()
	m.accessExports[export.DPDPRequestID] = export
	return nil
}

func (m *MemoryRepository) GetAccessExportByRequestID(ctx context.Context, requestID string) (*DPDPAccessExport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	export, ok := m.accessExports[requestID]
	if !ok {
		return nil, errors.New("access export not found for request")
	}
	return export, nil
}

func (m *MemoryRepository) UpdateAccessExportSections(ctx context.Context, export *DPDPAccessExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.accessExports[export.DPDPRequestID]
	if !ok {
		return errors.New("access export not found")
	}
	existing.ServicesReported = export.ServicesReported
	existing.Sections = export.Sections
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryRepository) UpdateAccessExportStatus(ctx context.Context, export *DPDPAccessExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.accessExports[export.DPDPRequestID]
	if !ok {
		return errors.New("access export not found")
	}
	existing.Status = export.Status
	existing.S3Key = export.S3Key
	existing.ReadyAt = export.ReadyAt
	existing.ExpiresAt = export.ExpiresAt
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryRepository) SweepExpiredAccessExports(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	count := 0
	for _, export := range m.accessExports {
		if export.Status != ExportStatusExpired && export.ExpiresAt != nil && now.After(*export.ExpiresAt) {
			export.Status = ExportStatusExpired
			export.UpdatedAt = now
			count++
		}
	}
	return count, nil
}

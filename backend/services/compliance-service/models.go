package main

import (
	"time"
)

// IRN Record Statuses
const (
	IRNStatusPending   = "PENDING"
	IRNStatusSubmitted = "SUBMITTED"
	IRNStatusIssued    = "ISSUED"
	IRNStatusFailed    = "FAILED"
	IRNStatusCancelled = "CANCELLED"
)

type IRNRecord struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	StoreID       string     `json:"store_id"`
	ChainID       string     `json:"chain_id"`
	Status        string     `json:"status"`
	IRN           *string    `json:"irn,omitempty"`
	AckNo         *string    `json:"ack_no,omitempty"`
	AckDate       *time.Time `json:"ack_date,omitempty"`
	SignedQRCode  *string    `json:"signed_qr_code,omitempty"`
	IRPPayload    string     `json:"irp_payload"`
	IRPResponse   *string    `json:"irp_response,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
	RetryCount    int        `json:"retry_count"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	IssuedAt      *time.Time `json:"issued_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type IRNOutboxEvent struct {
	ID          string     `json:"id"`
	Topic       string     `json:"topic"`
	Payload     string     `json:"payload"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

// DPDP Consent & Requests
const (
	DPDPStatusReceived   = "RECEIVED"
	DPDPStatusInProgress = "IN_PROGRESS"
	DPDPStatusCompleted  = "COMPLETED"
	DPDPStatusRejected   = "REJECTED"
)

type DPDPConsent struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	UserType       string     `json:"user_type"` // CUSTOMER | STAFF | CHAIN_HQ | ADMIN
	ConsentType    string     `json:"consent_type"` // MARKETING_COMMS | LOCATION_TRACKING | ANALYTICS_SHARING
	Granted        bool       `json:"granted"`
	GrantedAt      time.Time  `json:"granted_at"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	ConsentVersion string     `json:"consent_version"`
	CreatedAt      time.Time  `json:"created_at"`

	// Derived field for response
	NeedsReconfirmation bool `json:"needs_reconfirmation,omitempty"`
}

type DPDPRequest struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	UserType        string     `json:"user_type"` // CUSTOMER | STAFF | CHAIN_HQ | ADMIN
	RequestType     string     `json:"request_type"` // ACCESS | DELETION | CORRECTION
	Status          string     `json:"status"`
	Detail          *string    `json:"detail,omitempty"`
	HandledBy       *string    `json:"handled_by,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	SLADueAt        time.Time  `json:"sla_due_at"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type DPDPDeletionAudit struct {
	ID               string    `json:"id"`
	DPDPRequestID    string    `json:"dpdp_request_id"`
	ServiceName      string    `json:"service_name"`
	TablesAffected   []string  `json:"tables_affected"`
	RowsDeletedCount int       `json:"rows_deleted_count"`
	ExecutedAt       time.Time `json:"executed_at"`
}

// DPDP Access Export Statuses
const (
	ExportStatusAssembling = "ASSEMBLING"
	ExportStatusReady      = "READY"
	ExportStatusDownloaded = "DOWNLOADED"
	ExportStatusExpired    = "EXPIRED"
)

type DPDPAccessExport struct {
	ID               string     `json:"id"`
	DPDPRequestID    string     `json:"dpdp_request_id"`
	Status           string     `json:"status"` // ASSEMBLING | READY | DOWNLOADED | EXPIRED
	S3Key            *string    `json:"s3_key,omitempty"`
	ExpectedServices []string   `json:"expected_services"`
	ServicesReported []string   `json:"services_reported"`
	Sections         string     `json:"sections"` // JSON object map
	ReadyAt          *time.Time `json:"ready_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Derived field for presigned download URL
	DownloadURL *string `json:"download_url,omitempty"`
}

// Merchant KYC
type MerchantKYC struct {
	StoreID           string     `json:"store_id"`
	GSTIN             *string    `json:"gstin,omitempty"`
	GSTINVerified     bool       `json:"gstin_verified"`
	PANNumber         *string    `json:"pan_number,omitempty"`
	PANVerified       bool       `json:"pan_verified"`
	BankAccountLast4  *string    `json:"bank_account_last4,omitempty"`
	RazorpayAccountID *string    `json:"razorpay_account_id,omitempty"`
	KYCStatus         string     `json:"kyc_status"` // PENDING | VERIFIED | REJECTED
	KYCCompletedAt    *time.Time `json:"kyc_completed_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Velocity Alert
type VelocityAlert struct {
	ID         string     `json:"id"`
	StoreID    string     `json:"store_id"`
	AlertType  string     `json:"alert_type"` // UNUSUAL_TRANSACTION_VOLUME | UNUSUAL_TRANSACTION_VALUE | RAPID_REFUNDS
	Detail     string     `json:"detail"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Settlement Reconciliation Report
type SettlementReport struct {
	ID                     string    `json:"id"`
	ReportDate             string    `json:"report_date"` // YYYY-MM-DD
	TotalTransactions      int       `json:"total_transactions"`
	TotalAmountPaise       int64     `json:"total_amount_paise"`
	TotalSettledAmountPaise int64     `json:"total_settled_amount_paise"`
	DiscrepancyCount       int       `json:"discrepancy_count"`
	Discrepancies          string    `json:"discrepancies"` // JSON array string
	Status                 string    `json:"status"`        // COMPLETED | INCOMPLETE
	GeneratedAt            time.Time `json:"generated_at"`
}

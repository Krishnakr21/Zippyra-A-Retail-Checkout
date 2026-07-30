package main

import (
	"time"
)

type Category struct {
	ID        string      `json:"id"`
	ChainID   string      `json:"chain_id"`
	Name      string      `json:"name"`
	ParentID  *string     `json:"parent_id,omitempty"`
	SortOrder int         `json:"sort_order"`
	IsActive  bool        `json:"is_active"`
	Children  []*Category `json:"children,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type HsnGstRate struct {
	HSNCode        string  `json:"hsn_code"`
	GSTRatePercent float64 `json:"gst_rate_percent"`
	CessPercent    float64 `json:"cess_percent"`
	Description    string  `json:"description,omitempty"`
}

type Product struct {
	ID                    string     `json:"id"`
	StoreID               string     `json:"store_id"`
	ChainID               string     `json:"chain_id"`
	Barcode               string     `json:"barcode"`
	Name                  string     `json:"name"`
	Description           string     `json:"description,omitempty"`
	CategoryID            *string    `json:"category_id,omitempty"`
	PricePaise            int64      `json:"price_paise"`
	MRPPaise              int64      `json:"mrp_paise"`
	HSNCode               string     `json:"hsn_code"`
	GSTRatePercent        float64    `json:"gst_rate_percent"`
	IsActive              bool       `json:"is_active"`
	IsReturnable          bool       `json:"is_returnable"`
	ImageURL              string     `json:"image_url,omitempty"`
	ThumbnailURL          string     `json:"thumbnail_url,omitempty"`
	ImageProcessingStatus string     `json:"image_processing_status,omitempty"` // PENDING | PROCESSED | FAILED
	SyncSeq               int64      `json:"sync_seq"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type ImageProcessedWebhookRequest struct {
	S3RawKey     string `json:"s3_raw_key"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	FullURL      string `json:"full_url,omitempty"`
	Status       string `json:"status"` // PROCESSED | FAILED
	ErrorMessage string `json:"error_message,omitempty"`
}

type BarcodeLookupResponse struct {
	ID             string  `json:"id"`
	StoreID        string  `json:"store_id"`
	Barcode        string  `json:"barcode"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	PricePaise     int64   `json:"price_paise"`
	MRPPaise       int64   `json:"mrp_paise"`
	HSNCode        string  `json:"hsn_code"`
	GSTRatePercent float64 `json:"gst_rate_percent"`
	ImageURL       string  `json:"image_url,omitempty"`
	ThumbnailURL   string  `json:"thumbnail_url,omitempty"`
	IsReturnable   bool    `json:"is_returnable"`
}

type SearchResponse struct {
	Products []*Product `json:"products"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Source   string     `json:"source"` // "elasticsearch" | "postgres_fallback"
}

type CatalogSyncRequest struct {
	StoreID  string `json:"store_id"`
	SinceSeq int64  `json:"since_seq"`
}

type CatalogSyncResponse struct {
	Products   []*Product `json:"products"`
	DeletedIDs []string   `json:"deleted_ids"`
	NewMaxSeq  int64      `json:"new_max_seq"`
	HasMore    bool       `json:"has_more"`
}

type AdminProductCreateRequest struct {
	StoreID      string  `json:"store_id"`
	ChainID      string  `json:"chain_id"`
	Barcode      string  `json:"barcode"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	CategoryID   *string `json:"category_id,omitempty"`
	PricePaise   int64   `json:"price_paise"`
	MRPPaise     int64   `json:"mrp_paise"`
	HSNCode      string  `json:"hsn_code"`
	IsActive     *bool   `json:"is_active,omitempty"`
	IsReturnable *bool   `json:"is_returnable,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
}

type AdminProductUpdateRequest struct {
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	CategoryID   *string `json:"category_id,omitempty"`
	PricePaise   *int64  `json:"price_paise,omitempty"`
	MRPPaise     *int64  `json:"mrp_paise,omitempty"`
	HSNCode      string  `json:"hsn_code,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	IsReturnable *bool   `json:"is_returnable,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
}

type ImportRowError struct {
	RowNumber int    `json:"row_number"`
	Barcode   string `json:"barcode,omitempty"`
	Reason    string `json:"reason"`
}

type CatalogImportJob struct {
	ID            string            `json:"id"`
	StoreID       string            `json:"store_id"`
	ChainID       string            `json:"chain_id"`
	Status        string            `json:"status"` // PENDING | PROCESSING | COMPLETED | FAILED
	TotalRows     int               `json:"total_rows"`
	ProcessedRows int               `json:"processed_rows"`
	ErrorRows     []*ImportRowError `json:"error_rows"`
	CreatedAt     time.Time         `json:"created_at"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
}

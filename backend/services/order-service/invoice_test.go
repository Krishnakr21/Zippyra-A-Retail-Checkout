package main

import (
	"context"
	"testing"
	"time"

	"github.com/zippyra/backend/services/order-service/invoice"
)

func TestInvoicePDF_GenerationAndPaiseTotals(t *testing.T) {
	data := &invoice.InvoiceData{
		OrderID:       "ord-pdf-001",
		InvoiceNumber: 1001,
		OrderDate:     time.Now(),
		StoreName:     "Zippyra Indiranagar",
		StoreAddress:  "100ft Road, Bengaluru",
		StoreGSTIN:    "29ABCDE1234F1ZH",
		CustomerName:  "", // Unprovided -> "Walk-in Customer"
		CustomerPhone: "+919876543210",
		SupplyType:    "INTRA_STATE",
		Items: []invoice.InvoiceItem{
			{SKU: "SKU-001", Name: "Organic Milk 1L", Quantity: 2, UnitPrice: 75.0, Subtotal: 150.0, GSTRate: 18.0, CGSTAmount: 13.5, SGSTAmount: 13.5},
			{SKU: "SKU-002", Name: "Artisan Bread", Quantity: 1, UnitPrice: 50.0, Subtotal: 50.0, GSTRate: 18.0, CGSTAmount: 4.5, SGSTAmount: 4.5},
		},
		SubtotalPaise: 20000,
		DiscountPaise: 0,
		CGSTPaise:     1800,
		SGSTPaise:     1800,
		IGSTPaise:     0,
		TotalPaise:    23600,
		IRNEnabled:    true,
	}

	pdfBytes, err := invoice.RenderInvoicePDF(data)
	if err != nil {
		t.Fatalf("RenderInvoicePDF failed: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatalf("Expected non-empty PDF byte output")
	}
}

func TestInvoicePDF_InterStateIGSTLayout(t *testing.T) {
	data := &invoice.InvoiceData{
		OrderID:       "ord-igst-001",
		InvoiceNumber: 1002,
		OrderDate:     time.Now(),
		StoreName:     "Zippyra Mumbai",
		StoreAddress:  "Bandra West, Mumbai, MH",
		StoreGSTIN:    "27ABCDE1234F1ZH",
		CustomerName:  "Sunil Rao",
		CustomerPhone: "+919876543211",
		SupplyType:    "INTER_STATE",
		Items: []invoice.InvoiceItem{
			{SKU: "SKU-LUX-01", Name: "Luxury Leather Wallet", Quantity: 1, UnitPrice: 2000.0, Subtotal: 2000.0, GSTRate: 18.0, IGSTAmount: 360.0},
		},
		SubtotalPaise: 200000,
		DiscountPaise: 0,
		CGSTPaise:     0,
		SGSTPaise:     0,
		IGSTPaise:     36000,
		TotalPaise:    236000,
		IRNEnabled:    false,
	}

	pdfBytes, err := invoice.RenderInvoicePDF(data)
	if err != nil {
		t.Fatalf("RenderInvoicePDF failed for INTER_STATE supply: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatalf("Expected non-empty PDF byte output")
	}
}

func TestInvoiceService_Phase1AndPhase2Regeneration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresRepository(db)
	order := &Order{
		ID:            "ord-phase-001",
		PaymentID:     "pay-phase-001",
		UserID:        "usr-001",
		StoreID:       "store-001",
		SubtotalPaise: 10000,
		TotalPaise:    11800,
		CGSTPaise:     900,
		SGSTPaise:     900,
		SupplyType:    "INTRA_STATE",
		Status:        StatusCompleted,
		CreatedAt:     time.Now(),
	}
	_, err := db.Exec(`INSERT INTO orders (id, payment_id, user_id, store_id, items, subtotal_paise, discount_paise, cgst_paise, sgst_paise, igst_paise, total_paise, payment_method, supply_type, status, created_at)
	VALUES ('ord-phase-001', 'pay-phase-001', 'usr-001', 'store-001', '[]', 10000, 0, 900, 900, 0, 11800, 'UPI', 'INTRA_STATE', 'COMPLETED', CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("Failed to insert test order: %v", err)
	}

	svc := NewRealInvoiceService(repo)

	// Phase 1 Generation (immediate order.completed)
	err = svc.GenerateAndUploadInvoice(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("GenerateAndUploadInvoice failed: %v", err)
	}

	o1, _ := repo.GetOrderByID(context.Background(), order.ID)
	if o1.InvoiceS3Key == nil || *o1.InvoiceS3Key == "" {
		t.Fatalf("Expected InvoiceS3Key to be set in Phase 1")
	}

	// Phase 2 Regeneration (post compliance.irn_issued)
	irnHash := "irn_64char_hash_mock_ord-phase-001"
	ackNo := "ACK123456"
	ackDate := time.Now()
	signedQR := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	err = svc.RegenerateInvoiceWithIRN(context.Background(), order.ID, irnHash, ackNo, ackDate, signedQR)
	if err != nil {
		t.Fatalf("RegenerateInvoiceWithIRN failed: %v", err)
	}

	o2, _ := repo.GetOrderByID(context.Background(), order.ID)
	if o2.IRN == nil || *o2.IRN != irnHash {
		t.Fatalf("Expected IRN to be updated in Phase 2, got %v", o2.IRN)
	}
}

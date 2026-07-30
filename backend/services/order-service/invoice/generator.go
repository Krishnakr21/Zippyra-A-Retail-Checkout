package invoice

import (
	"context"
	"fmt"
	"log"
	"time"
)

type InvoiceRepository interface {
	GetOrderByID(ctx context.Context, id string) (interface{}, error)
	UpdateOrderInvoiceAndIRN(ctx context.Context, id string, invoiceS3Key, irn, ackNo *string, ackDate *time.Time, signedQR *string) error
}

type S3Uploader interface {
	UploadFile(ctx context.Context, s3Key string, content []byte, contentType string) error
	GetSignedURL(ctx context.Context, s3Key string, expirySeconds int) string
}

type DefaultInvoiceGenerator struct {
	repo     InvoiceRepository
	uploader S3Uploader
}

func NewDefaultInvoiceGenerator(repo InvoiceRepository, uploader S3Uploader) *DefaultInvoiceGenerator {
	return &DefaultInvoiceGenerator{
		repo:     repo,
		uploader: uploader,
	}
}

func (g *DefaultInvoiceGenerator) GenerateAndUpload(ctx context.Context, data *InvoiceData) (string, error) {
	pdfBytes, err := RenderInvoicePDF(data)
	if err != nil {
		return "", fmt.Errorf("failed to render invoice pdf for order %s: %w", data.OrderID, err)
	}

	year := time.Now().Format("2006")
	s3Key := fmt.Sprintf("invoices/%s/order_%s.pdf", year, data.OrderID)

	if g.uploader != nil {
		if err := g.uploader.UploadFile(ctx, s3Key, pdfBytes, "application/pdf"); err != nil {
			log.Printf("[Invoice Generator] Failed to upload PDF to S3 key %s: %v", s3Key, err)
			return "", err
		}
	}

	if g.repo != nil {
		_ = g.repo.UpdateOrderInvoiceAndIRN(ctx, data.OrderID, &s3Key, data.IRN, data.AckNo, data.AckDate, data.SignedQR)
	}

	log.Printf("[Invoice Generator] Successfully generated and uploaded PDF invoice for order %s (key: %s, size: %d bytes)", data.OrderID, s3Key, len(pdfBytes))
	return s3Key, nil
}

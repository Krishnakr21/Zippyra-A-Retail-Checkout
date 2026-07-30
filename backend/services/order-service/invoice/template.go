package invoice

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"

	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type InvoiceItem struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	Subtotal   float64 `json:"subtotal"`
	GSTRate    float64 `json:"gst_rate"`
	CGSTAmount float64 `json:"cgst_amount"`
	SGSTAmount float64 `json:"sgst_amount"`
	IGSTAmount float64 `json:"igst_amount"`
}

type InvoiceData struct {
	OrderID         string        `json:"order_id"`
	InvoiceNumber   int64         `json:"invoice_number"`
	OrderDate       time.Time     `json:"order_date"`
	StoreName       string        `json:"store_name"`
	StoreAddress    string        `json:"store_address"`
	StoreGSTIN      string        `json:"store_gstin"`
	CustomerName    string        `json:"customer_name"`
	CustomerPhone   string        `json:"customer_phone"`
	CustomerGSTIN   string        `json:"customer_gstin,omitempty"`
	SupplyType      string        `json:"supply_type"` // INTRA_STATE | INTER_STATE
	Items           []InvoiceItem `json:"items"`
	SubtotalPaise   int64         `json:"subtotal_paise"`
	DiscountPaise   int64         `json:"discount_paise"`
	CGSTPaise       int64         `json:"cgst_paise"`
	SGSTPaise       int64         `json:"sgst_paise"`
	IGSTPaise       int64         `json:"igst_paise"`
	TotalPaise      int64         `json:"total_paise"`

	// IRN E-Invoicing Block
	IRNEnabled bool       `json:"irn_enabled"`
	IRN        *string    `json:"irn,omitempty"`
	AckNo      *string    `json:"ack_no,omitempty"`
	AckDate    *time.Time `json:"ack_date,omitempty"`
	SignedQR   *string    `json:"signed_qr,omitempty"`
}

func RenderInvoicePDF(data *InvoiceData) ([]byte, error) {
	m := maroto.New()

	// 1. Header Row
	invNumStr := fmt.Sprintf("INV-%s-%06d", data.OrderDate.Format("20060102"), data.InvoiceNumber)
	if data.InvoiceNumber == 0 {
		invNumStr = fmt.Sprintf("INV-%s", data.OrderID)
	}

	m.AddRows(
		row.New(20).Add(
			col.New(6).Add(
				text.New("ZIPPYRA RETAIL", props.Text{
					Style: fontstyle.Bold,
					Size:  18,
					Align: align.Left,
				}),
			),
			col.New(6).Add(
				text.New("TAX INVOICE", props.Text{
					Style: fontstyle.Bold,
					Size:  18,
					Align: align.Right,
				}),
			),
		),
		row.New(12).Add(
			col.New(6).Add(
				text.New(fmt.Sprintf("Invoice #: %s", invNumStr), props.Text{Size: 10, Style: fontstyle.Bold}),
			),
			col.New(6).Add(
				text.New(fmt.Sprintf("Date: %s", data.OrderDate.Format("02 Jan 2006 15:04 MST")), props.Text{Size: 10, Align: align.Right}),
			),
		),
		row.New(10).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Order Ref ID: %s", data.OrderID), props.Text{Size: 9}),
			),
		),
	)

	// 2. Seller & Buyer Details Block
	custName := data.CustomerName
	if strings.TrimSpace(custName) == "" {
		custName = "Walk-in Customer"
	}

	m.AddRows(
		row.New(15).Add(
			col.New(6).Add(
				text.New("SELLER DETAILS", props.Text{Style: fontstyle.Bold, Size: 11}),
				text.New(fmt.Sprintf("Store: %s", data.StoreName), props.Text{Size: 9}),
				text.New(fmt.Sprintf("Address: %s", data.StoreAddress), props.Text{Size: 9}),
				text.New(fmt.Sprintf("GSTIN: %s", data.StoreGSTIN), props.Text{Size: 9, Style: fontstyle.Bold}),
			),
			col.New(6).Add(
				text.New("BUYER DETAILS", props.Text{Style: fontstyle.Bold, Size: 11}),
				text.New(fmt.Sprintf("Customer: %s", custName), props.Text{Size: 9}),
				text.New(fmt.Sprintf("Phone: %s", data.CustomerPhone), props.Text{Size: 9}),
				text.New(fmt.Sprintf("Buyer GSTIN: %s", defaultVal(&data.CustomerGSTIN, "N/A")), props.Text{Size: 9}),
			),
		),
	)

	// 3. Line Items Table Header
	isInterState := data.SupplyType == "INTER_STATE"

	if isInterState {
		m.AddRows(
			row.New(10).Add(
				col.New(4).Add(text.New("Item / Description", props.Text{Style: fontstyle.Bold, Size: 9})),
				col.New(2).Add(text.New("Qty", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Center})),
				col.New(2).Add(text.New("Unit Price", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
				col.New(2).Add(text.New("IGST Rate", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
				col.New(2).Add(text.New("Total (₹)", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
			),
		)
	} else {
		m.AddRows(
			row.New(10).Add(
				col.New(3).Add(text.New("Item / Description", props.Text{Style: fontstyle.Bold, Size: 9})),
				col.New(1).Add(text.New("Qty", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Center})),
				col.New(2).Add(text.New("Unit Price", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
				col.New(2).Add(text.New("CGST", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
				col.New(2).Add(text.New("SGST", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
				col.New(2).Add(text.New("Total (₹)", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right})),
			),
		)
	}

	// 4. Line Items Rows
	for _, item := range data.Items {
		if isInterState {
			m.AddRows(
				row.New(8).Add(
					col.New(4).Add(text.New(item.Name, props.Text{Size: 8})),
					col.New(2).Add(text.New(fmt.Sprintf("%d", item.Quantity), props.Text{Size: 8, Align: align.Center})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f", item.UnitPrice), props.Text{Size: 8, Align: align.Right})),
					col.New(2).Add(text.New(fmt.Sprintf("%.1f%%", item.GSTRate), props.Text{Size: 8, Align: align.Right})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f", item.Subtotal), props.Text{Size: 8, Align: align.Right})),
				),
			)
		} else {
			halfRate := item.GSTRate / 2.0
			m.AddRows(
				row.New(8).Add(
					col.New(3).Add(text.New(item.Name, props.Text{Size: 8})),
					col.New(1).Add(text.New(fmt.Sprintf("%d", item.Quantity), props.Text{Size: 8, Align: align.Center})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f", item.UnitPrice), props.Text{Size: 8, Align: align.Right})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f (%.1f%%)", item.CGSTAmount, halfRate), props.Text{Size: 8, Align: align.Right})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f (%.1f%%)", item.SGSTAmount, halfRate), props.Text{Size: 8, Align: align.Right})),
					col.New(2).Add(text.New(fmt.Sprintf("₹%.2f", item.Subtotal), props.Text{Size: 8, Align: align.Right})),
				),
			)
		}
	}

	// 5. Totals Block
	subtotalRs := float64(data.SubtotalPaise) / 100.0
	discountRs := float64(data.DiscountPaise) / 100.0
	cgstRs := float64(data.CGSTPaise) / 100.0
	sgstRs := float64(data.SGSTPaise) / 100.0
	igstRs := float64(data.IGSTPaise) / 100.0
	totalRs := float64(data.TotalPaise) / 100.0

	m.AddRows(
		row.New(12).Add(
			col.New(6).Add(text.New("SUMMARY", props.Text{Style: fontstyle.Bold, Size: 10})),
			col.New(6).Add(text.New(fmt.Sprintf("Subtotal: ₹%.2f", subtotalRs), props.Text{Size: 9, Align: align.Right})),
		),
	)

	if discountRs > 0 {
		m.AddRows(
			row.New(8).Add(
				col.New(12).Add(text.New(fmt.Sprintf("Discount: -₹%.2f", discountRs), props.Text{Size: 9, Align: align.Right})),
			),
		)
	}

	if isInterState {
		m.AddRows(
			row.New(8).Add(
				col.New(12).Add(text.New(fmt.Sprintf("Total IGST: ₹%.2f", igstRs), props.Text{Size: 9, Align: align.Right})),
			),
		)
	} else {
		m.AddRows(
			row.New(8).Add(
				col.New(12).Add(text.New(fmt.Sprintf("Total CGST: ₹%.2f | Total SGST: ₹%.2f", cgstRs, sgstRs), props.Text{Size: 9, Align: align.Right})),
			),
		)
	}

	m.AddRows(
		row.New(12).Add(
			col.New(12).Add(text.New(fmt.Sprintf("GRAND TOTAL: ₹%.2f", totalRs), props.Text{Style: fontstyle.Bold, Size: 12, Align: align.Right})),
		),
	)

	// 6. IRN E-Invoicing Block
	if data.IRNEnabled {
		if data.IRN != nil && *data.IRN != "" {
			ackDateStr := ""
			if data.AckDate != nil {
				ackDateStr = data.AckDate.Format("02 Jan 2006 15:04 MST")
			}

			qrContent := *data.IRN
			if data.SignedQR != nil && *data.SignedQR != "" {
				qrContent = *data.SignedQR
			}

			m.AddRows(
				row.New(12).Add(
					col.New(12).Add(text.New("E-INVOICE REFERENCE (IRN & IRP ACKNOWLEDGMENT)", props.Text{Style: fontstyle.Bold, Size: 10})),
				),
				row.New(10).Add(
					col.New(8).Add(
						text.New(fmt.Sprintf("IRN: %s", *data.IRN), props.Text{Size: 8}),
						text.New(fmt.Sprintf("Ack No: %s | Ack Date: %s", defaultVal(data.AckNo, "N/A"), ackDateStr), props.Text{Size: 8}),
					),
					col.New(4).Add(
						getQRComponent(qrContent),
					),
				),
			)
		} else {
			m.AddRows(
				row.New(10).Add(
					col.New(12).Add(text.New("E-invoice reference pending IRP generation", props.Text{Style: fontstyle.Italic, Size: 9})),
				),
			)
		}
	}

	// 7. Footer
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New("Return Policy: Returnable within 24 hours except non-returnable categories (see zippyra.com/returns)", props.Text{Size: 8, Align: align.Center}),
				text.New("Thank you for shopping with Zippyra! Support: support@zippyra.com", props.Text{Size: 8, Align: align.Center}),
			),
		),
	)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate maroto pdf document: %w", err)
	}

	return document.GetBytes(), nil
}

func defaultVal(ptr *string, fallback string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return fallback
}

func getQRComponent(content string) core.Component {
	if strings.HasPrefix(content, "data:image") || len(content) > 300 {
		// Base64 Image
		b64Data := content
		if idx := strings.Index(content, ","); idx != -1 {
			b64Data = content[idx+1:]
		}
		imgBytes, err := base64.StdEncoding.DecodeString(b64Data)
		if err == nil && len(imgBytes) > 0 {
			return image.NewFromBytes(imgBytes, ".png")
		}
	}
	// Fallback Code Component
	return code.NewBar(content)
}

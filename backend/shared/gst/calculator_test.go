package gst

import (
	"testing"
)

func TestCalculateGST_IntraState_And_InterState(t *testing.T) {
	items := []CartItemDTO{
		{Barcode: "8901030300011", Name: "Coffee", Qty: 2, PricePaise: 25000, HSNCode: "0901"},   // subtotal = 50000 (₹500.00)
		{Barcode: "012345678905", Name: "Biscuits", Qty: 1, PricePaise: 10000, HSNCode: "1905"}, // subtotal = 10000 (₹100.00)
	}

	hsnRates := map[string]float64{
		"0901": 5.0,  // 5% GST (2.5% CGST + 2.5% SGST)
		"1905": 18.0, // 18% GST (9% CGST + 9% SGST)
	}

	// 1. Same state customer (IntraState: State 27 -> State 27)
	intraRes := CalculateGST(items, "27", "27", 10000, hsnRates) // ₹100.00 discount

	if intraRes.SupplyType != "intrastate" {
		t.Errorf("Expected supplyType 'intrastate', got %s", intraRes.SupplyType)
	}

	if intraRes.SubtotalPaise != 60000 {
		t.Errorf("Expected subtotal 60000 paise, got %d", intraRes.SubtotalPaise)
	}
	if intraRes.DiscountPaise != 10000 {
		t.Errorf("Expected discount 10000 paise, got %d", intraRes.DiscountPaise)
	}
	if intraRes.TaxablePaise != 50000 {
		t.Errorf("Expected taxable 50000 paise, got %d", intraRes.TaxablePaise)
	}
	if intraRes.IGSTPaise != 0 {
		t.Errorf("Expected IGST 0 for intrastate, got %d", intraRes.IGSTPaise)
	}
	if intraRes.CGSTPaise <= 0 || intraRes.SGSTPaise <= 0 {
		t.Errorf("Expected positive CGST and SGST for intrastate, got CGST=%d, SGST=%d", intraRes.CGSTPaise, intraRes.SGSTPaise)
	}
	if intraRes.CGSTPaise != intraRes.SGSTPaise {
		t.Errorf("Expected CGST equal to SGST for intrastate, got CGST=%d, SGST=%d", intraRes.CGSTPaise, intraRes.SGSTPaise)
	}

	// 2. Cross state customer (InterState: State 27 store -> State 07 customer)
	interRes := CalculateGST(items, "27", "07", 10000, hsnRates)

	if interRes.SupplyType != "interstate" {
		t.Errorf("Expected supplyType 'interstate', got %s", interRes.SupplyType)
	}
	if interRes.CGSTPaise != 0 || interRes.SGSTPaise != 0 {
		t.Errorf("Expected CGST 0 and SGST 0 for interstate, got CGST=%d, SGST=%d", interRes.CGSTPaise, interRes.SGSTPaise)
	}
	if interRes.IGSTPaise <= 0 {
		t.Errorf("Expected positive IGST for interstate, got %d", interRes.IGSTPaise)
	}
}

func TestCalculateGST_PerItemRoundingAccuracy(t *testing.T) {
	// Multi-item cart designed to test rounding
	items := []CartItemDTO{
		{Barcode: "b1", Name: "Item 1", Qty: 1, PricePaise: 3333, HSNCode: "0901"},
		{Barcode: "b2", Name: "Item 2", Qty: 1, PricePaise: 3333, HSNCode: "0901"},
		{Barcode: "b3", Name: "Item 3", Qty: 1, PricePaise: 3334, HSNCode: "0901"},
	}

	hsnRates := map[string]float64{"0901": 5.0}

	res := CalculateGST(items, "27", "27", 0, hsnRates)

	var sumItemCGST int64
	var sumItemSGST int64
	for _, item := range res.ItemBreakdowns {
		sumItemCGST += item.CGSTPaise
		sumItemSGST += item.SGSTPaise
	}

	if sumItemCGST != res.CGSTPaise {
		t.Errorf("CGST rounding mismatch: item sum %d != total %d", sumItemCGST, res.CGSTPaise)
	}
	if sumItemSGST != res.SGSTPaise {
		t.Errorf("SGST rounding mismatch: item sum %d != total %d", sumItemSGST, res.SGSTPaise)
	}
	if res.TotalPaise != res.TaxablePaise+res.CGSTPaise+res.SGSTPaise+res.IGSTPaise {
		t.Errorf("TotalPaise mismatch: %d != %d + %d + %d + %d", res.TotalPaise, res.TaxablePaise, res.CGSTPaise, res.SGSTPaise, res.IGSTPaise)
	}
}

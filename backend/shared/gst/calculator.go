package gst

import (
	"math"
)

type CartItemDTO struct {
	Barcode        string `json:"barcode"`
	Name           string `json:"name"`
	Qty            int    `json:"qty"`
	PricePaise     int64  `json:"price_paise"`
	HSNCode        string `json:"hsn_code"`
	LineTotalPaise int64  `json:"line_total_paise"`
}

type ItemGSTBreakdown struct {
	Barcode        string  `json:"barcode"`
	TaxablePaise   int64   `json:"taxable_paise"`
	GSTRatePercent float64 `json:"gst_rate_percent"`
	CGSTPaise      int64   `json:"cgst_paise"`
	SGSTPaise      int64   `json:"sgst_paise"`
	IGSTPaise      int64   `json:"igst_paise"`
}

type GSTBreakdown struct {
	SubtotalPaise  int64              `json:"subtotal_paise"`
	DiscountPaise  int64              `json:"discount_paise"`
	TaxablePaise   int64              `json:"taxable_paise"`
	CGSTPaise      int64              `json:"cgst_paise"`
	SGSTPaise      int64              `json:"sgst_paise"`
	IGSTPaise      int64              `json:"igst_paise"`
	TotalPaise     int64              `json:"total_paise"`
	SupplyType     string             `json:"supply_type"` // "intrastate" | "interstate"
	ItemBreakdowns []ItemGSTBreakdown `json:"item_breakdowns"`
}

func CalculateGST(items []CartItemDTO, storeStateCode, customerGSTINStateCode string, discountPaise int64, hsnRates map[string]float64) GSTBreakdown {
	if storeStateCode == "" {
		storeStateCode = "27" // Default to Maharashtra if unconfigured
	}

	isIntraState := customerGSTINStateCode == "" || customerGSTINStateCode == storeStateCode

	supplyType := "intrastate"
	if !isIntraState {
		supplyType = "interstate"
	}

	var subtotalPaise int64
	for i := range items {
		if items[i].LineTotalPaise <= 0 {
			items[i].LineTotalPaise = items[i].PricePaise * int64(items[i].Qty)
		}
		subtotalPaise += items[i].LineTotalPaise
	}

	if discountPaise < 0 {
		discountPaise = 0
	}
	if discountPaise > subtotalPaise {
		discountPaise = subtotalPaise
	}

	taxablePaise := subtotalPaise - discountPaise

	var totalCGST, totalSGST, totalIGST int64
	var itemBreakdowns []ItemGSTBreakdown

	var allocatedDiscountSum int64

	for i, item := range items {
		lineTotal := item.LineTotalPaise

		var itemDiscount int64
		if subtotalPaise > 0 {
			if i == len(items)-1 {
				itemDiscount = discountPaise - allocatedDiscountSum
			} else {
				itemDiscount = (lineTotal * discountPaise) / subtotalPaise
				allocatedDiscountSum += itemDiscount
			}
		}

		itemTaxable := lineTotal - itemDiscount
		if itemTaxable < 0 {
			itemTaxable = 0
		}

		ratePercent := 18.0
		if hsnRates != nil {
			if r, ok := hsnRates[item.HSNCode]; ok && r >= 0 {
				ratePercent = r
			}
		}

		var cgst, sgst, igst int64
		if isIntraState {
			halfRate := ratePercent / 2.0
			cgst = int64(math.Round(float64(itemTaxable) * halfRate / 100.0))
			sgst = int64(math.Round(float64(itemTaxable) * halfRate / 100.0))
		} else {
			igst = int64(math.Round(float64(itemTaxable) * ratePercent / 100.0))
		}

		totalCGST += cgst
		totalSGST += sgst
		totalIGST += igst

		itemBreakdowns = append(itemBreakdowns, ItemGSTBreakdown{
			Barcode:        item.Barcode,
			TaxablePaise:   itemTaxable,
			GSTRatePercent: ratePercent,
			CGSTPaise:      cgst,
			SGSTPaise:      sgst,
			IGSTPaise:      igst,
		})
	}

	totalPaise := taxablePaise + totalCGST + totalSGST + totalIGST

	return GSTBreakdown{
		SubtotalPaise:  subtotalPaise,
		DiscountPaise:  discountPaise,
		TaxablePaise:   taxablePaise,
		CGSTPaise:      totalCGST,
		SGSTPaise:      totalSGST,
		IGSTPaise:      totalIGST,
		TotalPaise:     totalPaise,
		SupplyType:     supplyType,
		ItemBreakdowns: itemBreakdowns,
	}
}

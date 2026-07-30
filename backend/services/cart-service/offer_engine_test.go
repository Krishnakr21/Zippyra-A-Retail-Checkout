package main

import (
	"context"
	"testing"
)

func TestOfferEngine_AbsentKey_ReturnsZeroDiscountWithoutPanic(t *testing.T) {
	offerEngine := NewMemoryOfferEngine()
	ctx := context.Background()

	items := []*CartItem{
		{Barcode: "8901030300011", Name: "Coffee", Qty: 1, PricePaiseSnapshot: 25000},
	}

	// Store with NO configured offer rules
	discount, offers, err := offerEngine.EvaluateOffers(ctx, "store-no-offers", items, 25000)
	if err != nil {
		t.Fatalf("Unexpected error evaluating offers for store without rules: %v", err)
	}

	if discount != 0 {
		t.Errorf("Expected 0 discount for missing offer rules, got %d", discount)
	}
	if len(offers) != 0 {
		t.Errorf("Expected empty applied offers list, got %v", offers)
	}
}

func TestOfferEngine_FirstMatchWinsOrder(t *testing.T) {
	offerEngine := NewMemoryOfferEngine()
	ctx := context.Background()
	storeID := "store-rules-1"

	// Configure 2 stacking candidate rules: Rule 1 (10% off), Rule 2 (20% off)
	rules := []*OfferRule{
		{
			ID:                "rule-1",
			Type:              "PERCENT_OFF",
			Value:             10.0,
			AppliesTo:         "ALL",
			MinCartValuePaise: 1000,
		},
		{
			ID:                "rule-2",
			Type:              "PERCENT_OFF",
			Value:             20.0,
			AppliesTo:         "ALL",
			MinCartValuePaise: 1000,
		},
	}

	_ = offerEngine.SetStoreOfferRules(ctx, storeID, rules)

	items := []*CartItem{
		{Barcode: "8901030300011", Name: "Coffee", Qty: 1, PricePaiseSnapshot: 10000}, // subtotal 10000 (₹100)
	}

	// Should apply Rule 1 ONLY (First Match Wins -> 10% off = 1000 paise)
	discount, offers, err := offerEngine.EvaluateOffers(ctx, storeID, items, 10000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if discount != 1000 {
		t.Errorf("Expected first match rule discount 1000 paise (10%%), got %d", discount)
	}
	if len(offers) != 1 || offers[0] != "10% Off Cart" {
		t.Errorf("Expected offer name '10%% Off Cart', got %v", offers)
	}
}

func TestOfferEngine_RegressionDiscountMathTypes(t *testing.T) {
	offerEngine := NewMemoryOfferEngine()
	ctx := context.Background()

	t.Run("FLAT_OFF rule discount calculation", func(t *testing.T) {
		storeID := "store-flat-off"
		rules := []*OfferRule{
			{
				ID:                "rule-flat",
				Type:              "FLAT_OFF",
				Value:             1500.0, // ₹15 flat off
				AppliesTo:         "ALL",
				MinCartValuePaise: 5000,
			},
		}
		_ = offerEngine.SetStoreOfferRules(ctx, storeID, rules)

		items := []*CartItem{{Barcode: "8901030300011", Qty: 1, PricePaiseSnapshot: 10000}}
		discount, offers, err := offerEngine.EvaluateOffers(ctx, storeID, items, 10000)
		if err != nil || discount != 1500 || len(offers) != 1 {
			t.Errorf("expected 1500 paise flat off, got %d (%v)", discount, offers)
		}
	})

	t.Run("BOGO Buy 2 Get 1 Free calculation", func(t *testing.T) {
		storeID := "store-bogo"
		rules := []*OfferRule{
			{
				ID:        "rule-bogo",
				Type:      "BOGO",
				Value:     1.0,
				AppliesTo: "BARCODE_LIST",
				TargetIDs: []string{"8901030300011"},
			},
		}
		_ = offerEngine.SetStoreOfferRules(ctx, storeID, rules)

		items := []*CartItem{
			{Barcode: "8901030300011", Qty: 3, PricePaiseSnapshot: 5000}, // 3 items -> 1 free = 5000 discount
		}
		discount, offers, err := offerEngine.EvaluateOffers(ctx, storeID, items, 15000)
		if err != nil || discount != 5000 || len(offers) != 1 {
			t.Errorf("expected 5000 paise BOGO discount, got %d (%v)", discount, offers)
		}
	})

	t.Run("CATEGORY_PERCENT_OFF calculation", func(t *testing.T) {
		storeID := "store-cat-percent"
		rules := []*OfferRule{
			{
				ID:        "rule-cat",
				Type:      "CATEGORY_PERCENT_OFF",
				Value:     50.0, // 50% off category cat-beverages
				AppliesTo: "CATEGORY",
				TargetIDs: []string{"cat-beverages"},
			},
		}
		_ = offerEngine.SetStoreOfferRules(ctx, storeID, rules)

		items := []*CartItem{
			{Barcode: "8901030300011", CategoryID: "cat-beverages", Qty: 2, PricePaiseSnapshot: 4000}, // 8000 * 50% = 4000
			{Barcode: "8901030300028", CategoryID: "cat-snacks", Qty: 1, PricePaiseSnapshot: 2000},
		}
		discount, offers, err := offerEngine.EvaluateOffers(ctx, storeID, items, 10000)
		if err != nil || discount != 4000 || len(offers) != 1 {
			t.Errorf("expected 4000 paise category discount, got %d (%v)", discount, offers)
		}
	})
}


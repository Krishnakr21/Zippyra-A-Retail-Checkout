package validator

import (
	"testing"
	"time"
)

func TestValidateOfferConfig(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	// Valid EAN-13 barcode
	validBarcode := "8901030300011"

	t.Run("Valid PERCENT_OFF offer", func(t *testing.T) {
		res := ValidateOfferConfig(
			"PERCENT_OFF",
			"ALL",
			nil,
			map[string]interface{}{"percent": 20.0},
			&now,
			&future,
		)
		if !res.IsValid {
			t.Fatalf("expected valid, got error: %s (%s)", res.ErrMsg, res.ErrCode)
		}
	})

	t.Run("Reject PERCENT_OFF >= 100", func(t *testing.T) {
		res := ValidateOfferConfig(
			"PERCENT_OFF",
			"ALL",
			nil,
			map[string]interface{}{"percent": 100.0},
			&now,
			&future,
		)
		if res.IsValid {
			t.Fatalf("expected invalid for 100 percent, got valid")
		}
		if res.ErrCode != "PERCENT_OFF_INVALID" {
			t.Errorf("expected PERCENT_OFF_INVALID code, got %s", res.ErrCode)
		}
	})

	t.Run("FLAT_OFF high value warning", func(t *testing.T) {
		res := ValidateOfferConfig(
			"FLAT_OFF",
			"ALL",
			nil,
			map[string]interface{}{"flat_paise": 600000}, // ₹6,000 > ₹5,000
			&now,
			&future,
		)
		if !res.IsValid {
			t.Fatalf("expected valid for high flat_paise with warning, got error")
		}
		if len(res.Warnings) == 0 {
			t.Errorf("expected high-value warning for > ₹5,000")
		}
	})

	t.Run("Valid BOGO offer", func(t *testing.T) {
		res := ValidateOfferConfig(
			"BOGO",
			"BARCODE_LIST",
			[]string{validBarcode},
			map[string]interface{}{"buy_qty": 2, "get_qty": 1},
			&now,
			&future,
		)
		if !res.IsValid {
			t.Fatalf("expected valid BOGO offer, got error: %s", res.ErrMsg)
		}
	})

	t.Run("Reject invalid BOGO get_qty > buy_qty", func(t *testing.T) {
		res := ValidateOfferConfig(
			"BOGO",
			"BARCODE_LIST",
			[]string{validBarcode},
			map[string]interface{}{"buy_qty": 1, "get_qty": 2},
			&now,
			&future,
		)
		if res.IsValid {
			t.Fatalf("expected invalid BOGO when get_qty > buy_qty")
		}
	})

	t.Run("Reject invalid barcode checksum", func(t *testing.T) {
		res := ValidateOfferConfig(
			"PERCENT_OFF",
			"BARCODE_LIST",
			[]string{"8901030300019"}, // Invalid checksum
			map[string]interface{}{"percent": 15.0},
			&now,
			&future,
		)
		if res.IsValid {
			t.Fatalf("expected invalid barcode checksum error")
		}
		if res.ErrCode != "INVALID_BARCODE_CHECKSUM" {
			t.Errorf("expected INVALID_BARCODE_CHECKSUM, got %s", res.ErrCode)
		}
	})

	t.Run("Reject invalid schedule window active_until <= active_from", func(t *testing.T) {
		res := ValidateOfferConfig(
			"PERCENT_OFF",
			"ALL",
			nil,
			map[string]interface{}{"percent": 15.0},
			&future,
			&past, // active_until before active_from
		)
		if res.IsValid {
			t.Fatalf("expected invalid schedule window")
		}
		if res.ErrCode != "INVALID_SCHEDULE_WINDOW" {
			t.Errorf("expected INVALID_SCHEDULE_WINDOW, got %s", res.ErrCode)
		}
	})
}

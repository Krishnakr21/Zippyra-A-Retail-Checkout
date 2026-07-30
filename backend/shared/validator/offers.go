package validator

import (
	"fmt"
	"time"
)

type OfferValidationResult struct {
	IsValid  bool
	ErrCode  string
	ErrMsg   string
	Warnings []string
}

func ValidateOfferConfig(
	offerType string,
	appliesTo string,
	targetIDs []string,
	ruleConfig map[string]interface{},
	activeFrom *time.Time,
	activeUntil *time.Time,
) OfferValidationResult {
	var warnings []string

	// 1. Schedule Window Check
	if activeFrom != nil && activeUntil != nil {
		if !activeUntil.After(*activeFrom) {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "INVALID_SCHEDULE_WINDOW",
				ErrMsg:  "active_until must be strictly after active_from",
			}
		}
	}

	// 2. AppliesTo & TargetIDs Check
	switch appliesTo {
	case "ALL":
		// TargetIDs not required
	case "CATEGORY":
		if len(targetIDs) == 0 {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "INVALID_TARGET_IDS",
				ErrMsg:  "target_ids cannot be empty when applies_to is CATEGORY",
			}
		}
	case "BARCODE_LIST":
		if len(targetIDs) == 0 {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "INVALID_TARGET_IDS",
				ErrMsg:  "target_ids cannot be empty when applies_to is BARCODE_LIST",
			}
		}
		for _, b := range targetIDs {
			if !ValidateBarcode(b) {
				return OfferValidationResult{
					IsValid: false,
					ErrCode: "INVALID_BARCODE_CHECKSUM",
					ErrMsg:  fmt.Sprintf("target barcode %s failed EAN-13/UPC-A checksum validation", b),
				}
			}
		}
	default:
		return OfferValidationResult{
			IsValid: false,
			ErrCode: "INVALID_APPLIES_TO",
			ErrMsg:  fmt.Sprintf("invalid applies_to option: %s", appliesTo),
		}
	}

	// 3. Rule Config Check per Type
	switch offerType {
	case "PERCENT_OFF", "CATEGORY_PERCENT_OFF":
		percentVal, ok := getFloatVal(ruleConfig, "percent")
		if !ok || percentVal < 1 || percentVal > 90 {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "PERCENT_OFF_INVALID",
				ErrMsg:  "rule_config.percent must be between 1 and 90",
			}
		}

	case "FLAT_OFF":
		flatVal, ok := getInt64Val(ruleConfig, "flat_paise")
		if !ok || flatVal <= 0 {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "FLAT_OFF_INVALID",
				ErrMsg:  "rule_config.flat_paise must be greater than 0",
			}
		}
		if flatVal > 500000 { // ₹5,000
			warnings = append(warnings, fmt.Sprintf("High-value promotion warning: flat_paise %d exceeds ₹5,000 ceiling", flatVal))
		}

	case "BOGO":
		buyQty, ok1 := getIntVal(ruleConfig, "buy_qty")
		getQty, ok2 := getIntVal(ruleConfig, "get_qty")
		if !ok1 || !ok2 || buyQty < 1 || getQty < 1 || getQty > buyQty {
			return OfferValidationResult{
				IsValid: false,
				ErrCode: "BOGO_CONFIG_INVALID",
				ErrMsg:  "buy_qty >= 1, get_qty >= 1, and get_qty <= buy_qty are required for BOGO offers",
			}
		}

	default:
		return OfferValidationResult{
			IsValid: false,
			ErrCode: "INVALID_OFFER_TYPE",
			ErrMsg:  fmt.Sprintf("unsupported offer type: %s", offerType),
		}
	}

	return OfferValidationResult{
		IsValid:  true,
		Warnings: warnings,
	}
}

func getFloatVal(m map[string]interface{}, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	}
	return 0, false
}

func getInt64Val(m map[string]interface{}, key string) (int64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return val, true
	}
	return 0, false
}

func getIntVal(m map[string]interface{}, key string) (int, bool) {
	v, ok := getInt64Val(m, key)
	return int(v), ok
}

package validator

import (
	"testing"
)

func TestValidateBarcode_Checksums(t *testing.T) {
	tests := []struct {
		barcode string
		expected bool
	}{
		{"8901030300011", true},  // Valid EAN-13
		{"4006381333931", true},  // Valid EAN-13
		{"8901030300028", true},  // Valid EAN-13
		{"8901030300019", false}, // Invalid EAN-13
		{"012345678905", true},   // Valid UPC-A
		{"036000291452", true},   // Valid UPC-A
		{"012345678909", false},  // Invalid UPC-A
		{"12345", false},         // Invalid length
		{"ABC1234567890", false}, // Non-numeric
	}

	for _, tt := range tests {
		res := ValidateBarcode(tt.barcode)
		if res != tt.expected {
			t.Errorf("ValidateBarcode(%s) = %v; expected %v", tt.barcode, res, tt.expected)
		}
	}
}

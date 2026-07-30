package validator

import (
	"testing"
)

func TestValidateGSTIN_ValidChecksum(t *testing.T) {
	validGSTINs := []string{
		"29ABCDE1234F1ZW",
		"07AAAAA0000A1Z4",
		"27AAPCU1427M1ZT",
	}

	for _, gstin := range validGSTINs {
		valid, err := ValidateGSTIN(gstin)
		if !valid || err != nil {
			t.Errorf("Expected GSTIN %s to be valid, got valid=%v, err=%v", gstin, valid, err)
		}
	}
}

func TestValidateGSTIN_SingleCharacterMutationFails(t *testing.T) {
	// Mutated check digit 'W' -> 'X'
	mutatedGSTIN := "29ABCDE1234F1ZX"
	valid, err := ValidateGSTIN(mutatedGSTIN)
	if valid || err == nil {
		t.Errorf("Expected mutated GSTIN %s to fail checksum, got valid=%v", mutatedGSTIN, valid)
	}

	// Mutated internal digit '2' -> '3' at position 0
	mutatedGSTIN2 := "39ABCDE1234F1ZW"
	valid2, err2 := ValidateGSTIN(mutatedGSTIN2)
	if valid2 || err2 == nil {
		t.Errorf("Expected mutated GSTIN %s to fail checksum, got valid=%v", mutatedGSTIN2, valid2)
	}
}

func TestValidateGSTIN_InvalidFormat(t *testing.T) {
	invalidFormats := []string{
		"INVALID_GSTIN",
		"29ABCDE1234F1Z",   // Too short (14 chars)
		"29ABCDE1234F1ZWW", // Too long (16 chars)
		"XXABCDE1234F1ZW",  // Letters instead of 2-digit state code
	}

	for _, gstin := range invalidFormats {
		valid, err := ValidateGSTIN(gstin)
		if valid || err == nil {
			t.Errorf("Expected invalid format GSTIN %s to fail, got valid=%v", gstin, valid)
		}
	}
}

func TestGSTINStateCode(t *testing.T) {
	tests := []struct {
		gstin    string
		expected string
	}{
		{"29ABCDE1234F1ZW", "29"},
		{"07AAAAA0000A1Z5", "07"},
		{"27AAPCU1427M1ZP", "27"},
		{"", ""},
		{"9", "9"},
	}

	for _, tt := range tests {
		actual := GSTINStateCode(tt.gstin)
		if actual != tt.expected {
			t.Errorf("GSTINStateCode(%s) = %s, expected %s", tt.gstin, actual, tt.expected)
		}
	}
}

package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const gstinAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var gstinRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)

var (
	ErrGSTINInvalidFormat   = errors.New("GSTIN format is invalid")
	ErrGSTINInvalidChecksum = errors.New("GSTIN checksum is invalid")
)

// ValidateGSTIN checks both regex format and GSTN Mod-36 checksum
func ValidateGSTIN(gstin string) (bool, error) {
	gstin = strings.TrimSpace(strings.ToUpper(gstin))

	if !gstinRegex.MatchString(gstin) {
		return false, ErrGSTINInvalidFormat
	}

	// GSTN Mod-36 checksum algorithm over first 14 characters
	sum := 0
	for i := 0; i < 14; i++ {
		char := rune(gstin[i])
		charVal := strings.IndexRune(gstinAlphabet, char)
		if charVal == -1 {
			return false, ErrGSTINInvalidFormat
		}

		// Alternating weights: index 0 (1st pos) weight 1, index 1 (2nd pos) weight 2, etc.
		factor := 1
		if i%2 != 0 {
			factor = 2
		}

		product := charVal * factor
		quotient := product / 36
		remainder := product % 36
		sum += quotient + remainder
	}

	checkVal := (36 - (sum % 36)) % 36
	expectedCheckChar := gstinAlphabet[checkVal]

	if gstin[14] != expectedCheckChar {
		return false, fmt.Errorf("%w: expected '%c', got '%c'", ErrGSTINInvalidChecksum, expectedCheckChar, gstin[14])
	}

	return true, nil
}

// GSTINStateCode returns the first 2 digits of the GSTIN (state code)
func GSTINStateCode(gstin string) string {
	gstin = strings.TrimSpace(gstin)
	if len(gstin) < 2 {
		return gstin
	}
	return gstin[:2]
}

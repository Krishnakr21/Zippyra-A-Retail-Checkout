package validator

import (
	"strings"
)

// ValidateEAN13 checks if barcode is a valid 13-digit EAN-13 string with valid modulo-10 checksum.
func ValidateEAN13(barcode string) bool {
	barcode = strings.TrimSpace(barcode)
	if len(barcode) != 13 {
		return false
	}
	for i := 0; i < 13; i++ {
		if barcode[i] < '0' || barcode[i] > '9' {
			return false
		}
	}

	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(barcode[i] - '0')
		if i%2 == 0 {
			sum += digit * 1 // 1st, 3rd, 5th, ... positions (0-indexed even)
		} else {
			sum += digit * 3 // 2nd, 4th, 6th, ... positions (0-indexed odd)
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	expectedCheckDigit := int(barcode[12] - '0')
	return checkDigit == expectedCheckDigit
}

// ValidateUPCA checks if barcode is a valid 12-digit UPC-A string with valid modulo-10 checksum.
func ValidateUPCA(barcode string) bool {
	barcode = strings.TrimSpace(barcode)
	if len(barcode) != 12 {
		return false
	}
	for i := 0; i < 12; i++ {
		if barcode[i] < '0' || barcode[i] > '9' {
			return false
		}
	}

	sum := 0
	for i := 0; i < 11; i++ {
		digit := int(barcode[i] - '0')
		if i%2 == 0 {
			sum += digit * 3 // 1st, 3rd, 5th, ... positions (0-indexed even)
		} else {
			sum += digit * 1 // 2nd, 4th, 6th, ... positions (0-indexed odd)
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	expectedCheckDigit := int(barcode[11] - '0')
	return checkDigit == expectedCheckDigit
}

// ValidateBarcode returns true if barcode is a valid EAN-13 or UPC-A barcode.
func ValidateBarcode(barcode string) bool {
	return ValidateEAN13(barcode) || ValidateUPCA(barcode)
}

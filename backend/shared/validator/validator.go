package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+91[6-9]\d{9}$`)

// ValidatePhone checks if phone matches ^\+91[6-9]\d{9}$
func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(strings.TrimSpace(phone))
}

// ValidateEmail checks if email is valid
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

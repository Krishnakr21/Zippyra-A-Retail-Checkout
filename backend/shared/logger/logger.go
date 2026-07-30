package logger

import (
	"log"
	"strings"
)

// MaskPhone masks phone number e.g. +919999999999 -> +91XXXXXX9999
func MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return "***"
	}
	prefix := phone[:3]
	suffix := phone[len(phone)-4:]
	return prefix + "XXXXXX" + suffix
}

// MaskEmail masks email e.g. user@example.com -> u***@example.com
func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***@***"
	}
	name := parts[0]
	domain := parts[1]
	maskedName := string(name[0]) + "***"
	return maskedName + "@" + domain
}

// Info logs masked information
func Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

// Warn logs warning information
func Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

// Debug logs debug information
func Debug(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

// Error logs error information
func Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

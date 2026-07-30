package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfig_MissingRequiredVars_AggregatedError(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("AWS_REGION", "ap-south-1")
	os.Setenv("DATABASE_URL", "")
	os.Setenv("JWT_PRIVATE_KEY_CURRENT", "")
	os.Setenv("JWT_SECRET", "")

	_, err := Load("auth-service")
	if err == nil {
		t.Fatalf("expected error when required vars missing, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "DATABASE_URL") {
		t.Errorf("expected error message to contain DATABASE_URL, got: %s", errStr)
	}
	if !strings.Contains(errStr, "JWT_PRIVATE_KEY_CURRENT") {
		t.Errorf("expected error message to contain JWT_PRIVATE_KEY_CURRENT, got: %s", errStr)
	}
}

func TestConfig_ValidConfig_Success(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("AWS_REGION", "ap-south-1")
	os.Setenv("SMS_PROVIDER", "log")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "secret-key-32bytes-long-string")
	os.Setenv("JWT_KID_CURRENT", "v1")

	cfg, err := Load("auth-service")
	if err != nil {
		t.Fatalf("unexpected error loading valid config: %v", err)
	}

	if cfg.ServiceName != "auth-service" {
		t.Errorf("expected serviceName 'auth-service', got '%s'", cfg.ServiceName)
	}
}

func TestConfig_ProductionSafetyInvariants_Violation(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("AWS_REGION", "ap-south-1")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	os.Setenv("JWT_SECRET", "secret-key-32bytes-long-string")
	os.Setenv("JWT_KID_CURRENT", "v1")
	os.Setenv("DB_SSL_MODE", "disable") // Violation!
	os.Setenv("RDS_IAM_AUTH_ENABLED", "false") // Violation!
	os.Setenv("WAF_ENABLED", "false") // Violation!

	_, err := Load("auth-service")
	if err == nil {
		t.Fatalf("expected error when prod safety invariants are violated, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "PROD SAFETY") {
		t.Errorf("expected PROD SAFETY violation error, got: %s", errStr)
	}
}

func TestConfig_DPDP_InvalidRegion(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("AWS_REGION", "us-east-1") // Violation! Must be ap-south-1 or ap-south-2

	_, err := Load("auth-service")
	if err == nil {
		t.Fatalf("expected error for non-Indian AWS region due to DPDP data localization")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "DPDP compliance") {
		t.Errorf("expected DPDP compliance error, got: %s", errStr)
	}
}

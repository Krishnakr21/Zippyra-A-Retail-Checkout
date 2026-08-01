package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_MasksSecrets(t *testing.T) {
	secret1 := "zip_agent_secret_key_123"
	secret2 := "whsec_super_secret_999"

	l, err := NewLogger("", secret1, secret2)
	if err != nil {
		t.Fatalf("unexpected error creating logger: %v", err)
	}

	var buf bytes.Buffer
	l.out = &buf

	l.Info("Connecting with API Key %s and Webhook secret %s", secret1, secret2)

	output := buf.String()
	if strings.Contains(output, secret1) || strings.Contains(output, secret2) {
		t.Errorf("Logger leaked secret! Output: %s", output)
	}

	if !strings.Contains(output, "[REDACTED_SECRET]") {
		t.Errorf("Logger failed to insert [REDACTED_SECRET] placeholder! Output: %s", output)
	}
}

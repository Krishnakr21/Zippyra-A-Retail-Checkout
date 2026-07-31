package main

import (
	"context"
	"testing"
)

func TestDPDPConsentService_VersionChecking(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewDPDPConsentService(repo)

	ctx := context.Background()

	// 1. Consent with current version v1.0
	c1, err := svc.RecordConsent(ctx, "usr-consent-1", "CUSTOMER", "MARKETING_COMMS", true, "v1.0")
	if err != nil {
		t.Fatalf("RecordConsent v1.0 failed: %v", err)
	}
	if c1.NeedsReconfirmation {
		t.Fatalf("v1.0 consent should not need reconfirmation")
	}

	// 2. Outdated consent version v0.9 -> needs reconfirmation
	c2, err := svc.RecordConsent(ctx, "usr-consent-1", "CUSTOMER", "LOCATION_TRACKING", true, "v0.9")
	if err != nil {
		t.Fatalf("RecordConsent v0.9 failed: %v", err)
	}
	if !c2.NeedsReconfirmation {
		t.Fatalf("Outdated v0.9 consent must set needs_reconfirmation = true")
	}

	// 3. Fetch consents list
	consents, err := svc.GetUserConsents(ctx, "usr-consent-1")
	if err != nil || len(consents) != 2 {
		t.Fatalf("Expected 2 consents for user, got %d", len(consents))
	}
}

package main

import (
	"context"
	"testing"
)

func TestDPDPConsentVersionBump_TriggersNeedsReconfirmation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewDPDPConsentService(repo)

	ctx := context.Background()
	userID := "user-consent-version-test-01"

	// 1. Record user consent under old version "v1.0"
	_, err := svc.RecordConsent(ctx, userID, "CUSTOMER", "MARKETING_COMMS", true, "v1.0")
	if err != nil {
		t.Fatalf("RecordConsent v1.0 failed: %v", err)
	}

	// 2. Fetch consents under CurrentConsentVersion ("v1.0" -> false)
	consentsV1, err := svc.GetUserConsents(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserConsents failed: %v", err)
	}
	if len(consentsV1) != 1 {
		t.Fatalf("Expected 1 consent record, got %d", len(consentsV1))
	}
	if consentsV1[0].NeedsReconfirmation {
		t.Errorf("Expected NeedsReconfirmation=false when version matches CurrentConsentVersion, got true")
	}

	// 3. Set DPDP_PRIVACY_NOTICE_VERSION env var to "v1.2" (simulating live version bump)
	t.Setenv("DPDP_PRIVACY_NOTICE_VERSION", "v1.2")

	// 4. Record older consent under "v1.0" for a second toggle
	_, err = svc.RecordConsent(ctx, userID, "CUSTOMER", "LOCATION_TRACKING", true, "v1.0")
	if err != nil {
		t.Fatalf("RecordConsent location v1.0 failed: %v", err)
	}

	// 5. Fetch consents and verify old v1.0 consent correctly triggers NeedsReconfirmation=true
	consentsUpdated, err := svc.GetUserConsents(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserConsents updated failed: %v", err)
	}

	for _, c := range consentsUpdated {
		if c.ConsentType == "LOCATION_TRACKING" {
			if !c.NeedsReconfirmation {
				t.Errorf("Expected NeedsReconfirmation=true for LOCATION_TRACKING recorded under v1.0 when notice version is v1.2")
			}
		}
	}
}

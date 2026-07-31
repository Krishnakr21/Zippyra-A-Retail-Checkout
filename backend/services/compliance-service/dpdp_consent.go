package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

const DefaultConsentVersion = "v1.0"

func GetActiveConsentVersion() string {
	if v := os.Getenv("DPDP_PRIVACY_NOTICE_VERSION"); v != "" {
		return v
	}
	return DefaultConsentVersion
}

type DPDPConsentService struct {
	repo Repository
}

func NewDPDPConsentService(repo Repository) *DPDPConsentService {
	return &DPDPConsentService{repo: repo}
}

func (s *DPDPConsentService) RecordConsent(ctx context.Context, userID, userType, consentType string, granted bool, version string) (*DPDPConsent, error) {
	if consentType == "" {
		return nil, fmt.Errorf("consent_type is required")
	}
	if userType == "" {
		userType = "CUSTOMER"
	}
	activeVersion := GetActiveConsentVersion()
	if version == "" {
		version = activeVersion
	}

	now := time.Now().UTC()
	var revokedAt *time.Time
	if !granted {
		revokedAt = &now
	}

	consent := &DPDPConsent{
		UserID:         userID,
		UserType:       userType,
		ConsentType:    consentType,
		Granted:        granted,
		GrantedAt:      now,
		RevokedAt:      revokedAt,
		ConsentVersion: version,
	}

	if err := s.repo.UpsertConsent(ctx, consent); err != nil {
		return nil, fmt.Errorf("failed to record consent: %w", err)
	}

	if version != activeVersion {
		consent.NeedsReconfirmation = true
	}
	return consent, nil
}

func (s *DPDPConsentService) GetUserConsents(ctx context.Context, userID string) ([]*DPDPConsent, error) {
	consents, err := s.repo.GetLatestConsentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user consents: %w", err)
	}

	activeVersion := GetActiveConsentVersion()
	for _, c := range consents {
		if c.ConsentVersion != activeVersion {
			c.NeedsReconfirmation = true
		}
	}
	return consents, nil
}

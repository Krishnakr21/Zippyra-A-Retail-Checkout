package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/validator"
)

type KYCService struct {
	repo Repository
}

func NewKYCService(repo Repository) *KYCService {
	return &KYCService{repo: repo}
}

func (s *KYCService) SubmitKYC(ctx context.Context, storeID, gstin, pan, bankLast4, rzpAccountID string) (*MerchantKYC, error) {
	if storeID == "" {
		return nil, fmt.Errorf("store_id is required")
	}

	gstinValid := false
	if gstin != "" {
		v, err := validator.ValidateGSTIN(gstin)
		if err == nil && v {
			gstinValid = true
		}
	}

	panValid := false
	if pan != "" && len(pan) == 10 {
		panValid = true
	}

	status := "PENDING"
	now := time.Now().UTC()
	var completedAt *time.Time

	if gstinValid && panValid && bankLast4 != "" {
		status = "VERIFIED"
		completedAt = &now
	}

	var gPtr, pPtr, bPtr, rPtr *string
	if gstin != "" { gPtr = &gstin }
	if pan != "" { pPtr = &pan }
	if bankLast4 != "" { bPtr = &bankLast4 }
	if rzpAccountID != "" { rPtr = &rzpAccountID }

	kyc := &MerchantKYC{
		StoreID:           storeID,
		GSTIN:             gPtr,
		GSTINVerified:     gstinValid,
		PANNumber:         pPtr,
		PANVerified:       panValid,
		BankAccountLast4:  bPtr,
		RazorpayAccountID: rPtr,
		KYCStatus:         status,
		KYCCompletedAt:    completedAt,
		UpdatedAt:         now,
	}

	if err := s.repo.UpsertMerchantKYC(ctx, kyc); err != nil {
		return nil, fmt.Errorf("failed to save KYC record: %w", err)
	}

	return kyc, nil
}

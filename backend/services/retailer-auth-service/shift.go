package main

import (
	"context"
	"fmt"
	"time"
)

type ShiftService struct {
	repo Repository
}

func NewShiftService(repo Repository) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) StartShift(ctx context.Context, staffID, storeID string) (*StaffShift, error) {
	shift := &StaffShift{
		StaffID: staffID,
		StoreID: storeID,
	}

	if err := s.repo.StartShiftTx(ctx, shift); err != nil {
		if err == ErrShiftActive {
			existing, _ := s.repo.GetActiveShift(ctx, staffID)
			startedAt := time.Now()
			if existing != nil {
				startedAt = existing.StartedAt
			}
			return nil, fmt.Errorf("%s: shift started at %s", CodeShiftAlreadyActive, startedAt.Format(time.RFC3339))
		}
		return nil, err
	}

	return shift, nil
}

func (s *ShiftService) EndShift(ctx context.Context, staffID string) error {
	if err := s.repo.EndShift(ctx, staffID); err != nil {
		if err == ErrNoActiveShift {
			return fmt.Errorf(CodeNoActiveShift)
		}
		return err
	}
	return nil
}

func (s *ShiftService) GetCurrentShift(ctx context.Context, staffID string) (*StaffShift, error) {
	return s.repo.GetActiveShift(ctx, staffID)
}

func (s *ShiftService) GetShiftHistory(ctx context.Context, storeID string, dateFrom, dateTo *time.Time, page, pageSize int) ([]*StaffShift, int64, error) {
	return s.repo.GetShiftHistory(ctx, storeID, dateFrom, dateTo, page, pageSize)
}

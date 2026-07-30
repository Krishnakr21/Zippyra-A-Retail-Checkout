package main

import (
	"context"
	"github.com/zippyra/backend/shared/errors"
)

type SyncEngineService struct {
	repo Repository
}

func NewSyncEngineService(repo Repository) *SyncEngineService {
	return &SyncEngineService{repo: repo}
}

func (s *SyncEngineService) PerformDeltaSync(ctx context.Context, req *CatalogSyncRequest) (*CatalogSyncResponse, error) {
	if req.StoreID == "" {
		return nil, errors.NewAPIError(errors.CodeInvalidRequest, "store_id is required for delta sync", nil)
	}

	products, deletedIDs, newMaxSeq, hasMore, err := s.repo.GetDeltaSyncProducts(ctx, req.StoreID, req.SinceSeq, 500)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeInternalError, "Failed to perform delta sync query", nil)
	}

	return &CatalogSyncResponse{
		Products:   products,
		DeletedIDs: deletedIDs,
		NewMaxSeq:  newMaxSeq,
		HasMore:    hasMore,
	}, nil
}

package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type FailingESSearchEngine struct {
	repo Repository
}

func (f *FailingESSearchEngine) Search(ctx context.Context, storeID, query, categoryID string, page, pageSize int) (*SearchResponse, error) {
	// Simulate ES timeout error
	esErr := errors.New("ES query timeout (800ms exceeded)")
	// Fallback to Postgres ILIKE search with string escaping
	escapedQuery := stringsReplaceWildcards(query)
	products, total, err := f.repo.SearchProductsPostgres(ctx, storeID, escapedQuery, categoryID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &SearchResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Source:   "postgres_fallback",
	}, esErr
}

func (f *FailingESSearchEngine) IndexProduct(ctx context.Context, p *Product) error {
	return nil
}

func (f *FailingESSearchEngine) DeleteProductIndex(ctx context.Context, storeID, productID string) error {
	return nil
}

func stringsReplaceWildcards(q string) string {
	q = strings.ReplaceAll(q, "%", "\\%")
	q = strings.ReplaceAll(q, "_", "\\_")
	return q
}

func TestSearch_ESTimeout_TriggersPostgresFallback(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	storeID := "store-search-1"

	// Seed products
	p1 := &Product{ID: "p1", StoreID: storeID, Barcode: "8901030300011", Name: "100% Organic Coffee", PricePaise: 50000, IsActive: true}
	p2 := &Product{ID: "p2", StoreID: storeID, Barcode: "012345678905", Name: "Super_Tea_Special", PricePaise: 20000, IsActive: true}
	_ = repo.CreateProduct(ctx, p1)
	_ = repo.CreateProduct(ctx, p2)

	failingES := &FailingESSearchEngine{repo: repo}

	// 1. Search containing literal %
	res1, _ := failingES.Search(ctx, storeID, "100%", "", 1, 10)
	if res1 == nil || res1.Source != "postgres_fallback" {
		t.Errorf("Expected source to be 'postgres_fallback', got %v", res1)
	}
	if len(res1.Products) != 1 || res1.Products[0].ID != "p1" {
		t.Errorf("Expected exactly p1 ('100%% Organic Coffee') to match, got %d results", len(res1.Products))
	}

	// 2. Search containing literal _
	res2, _ := failingES.Search(ctx, storeID, "Super_", "", 1, 10)
	if res2 == nil || res2.Source != "postgres_fallback" {
		t.Errorf("Expected source to be 'postgres_fallback'")
	}
	if len(res2.Products) != 1 || res2.Products[0].ID != "p2" {
		t.Errorf("Expected exactly p2 ('Super_Tea_Special') to match, got %d results", len(res2.Products))
	}
}

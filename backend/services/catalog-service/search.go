package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type SearchEngine interface {
	Search(ctx context.Context, storeID, query, categoryID string, page, pageSize int) (*SearchResponse, error)
	IndexProduct(ctx context.Context, p *Product) error
	DeleteProductIndex(ctx context.Context, storeID, productID string) error
}

type ESSearchEngine struct {
	esEndpoint string
	repo       Repository
	client     *http.Client
}

func NewESSearchEngine(esEndpoint string, repo Repository) SearchEngine {
	if esEndpoint == "" {
		esEndpoint = "http://localhost:9200"
	}
	return &ESSearchEngine{
		esEndpoint: esEndpoint,
		repo:       repo,
		client:     &http.Client{Timeout: 800 * time.Millisecond}, // 800ms strict timeout
	}
}

func (e *ESSearchEngine) Search(ctx context.Context, storeID, query, categoryID string, page, pageSize int) (*SearchResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	from := (page - 1) * pageSize

	// 1. Attempt ES Multi-Match Query (800ms context timeout)
	esCtx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	resp, err := e.queryElasticsearch(esCtx, storeID, query, categoryID, from, pageSize)
	if err == nil && resp != nil {
		resp.Page = page
		resp.PageSize = pageSize
		resp.Source = "elasticsearch"
		return resp, nil
	}

	// 2. On ES error/timeout, trigger Postgres ILIKE fallback
	logger.Warn("Elasticsearch query failed or timed out (%v). Falling back to Postgres ILIKE search for query '%s'", err, query)
	return e.postgresFallbackSearch(ctx, storeID, query, categoryID, page, pageSize)
}

func (e *ESSearchEngine) queryElasticsearch(ctx context.Context, storeID, query, categoryID string, from, size int) (*SearchResponse, error) {
	if e.esEndpoint == "" {
		return nil, fmt.Errorf("ES endpoint unconfigured")
	}

	mustClauses := []map[string]interface{}{
		{"term": map[string]interface{}{"store_id": storeID}},
		{"term": map[string]interface{}{"is_active": true}},
		{"term": map[string]interface{}{"is_deleted": false}},
	}

	if categoryID != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{"category_id": categoryID},
		})
	}

	if query != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name^3", "category_name^2", "barcode"},
			},
		})
	}

	reqBody := map[string]interface{}{
		"from": from,
		"size": size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/zippyra_products/_search", strings.TrimRight(e.esEndpoint, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ES returned status code %d", httpResp.StatusCode)
	}

	var esResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&esResult); err != nil {
		return nil, err
	}

	var products []*Product
	for _, hit := range esResult.Hits.Hits {
		p := hit.Source
		products = append(products, &p)
	}

	return &SearchResponse{
		Products: products,
		Total:    esResult.Hits.Total.Value,
	}, nil
}

func (e *ESSearchEngine) postgresFallbackSearch(ctx context.Context, storeID, query, categoryID string, page, pageSize int) (*SearchResponse, error) {
	// Escape wildcard characters % and _ for safe ILIKE search
	escapedQuery := strings.ReplaceAll(query, "%", "\\%")
	escapedQuery = strings.ReplaceAll(escapedQuery, "_", "\\_")

	products, total, err := e.repo.SearchProductsPostgres(ctx, storeID, escapedQuery, categoryID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &SearchResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Source:   "postgres_fallback",
	}, nil
}

func (e *ESSearchEngine) IndexProduct(ctx context.Context, p *Product) error {
	if e.esEndpoint == "" || p == nil {
		return nil
	}

	doc := map[string]interface{}{
		"id":             p.ID,
		"store_id":       p.StoreID,
		"chain_id":       p.ChainID,
		"barcode":        p.Barcode,
		"name":           p.Name,
		"description":    p.Description,
		"category_id":    p.CategoryID,
		"price_paise":    p.PricePaise,
		"mrp_paise":      p.MRPPaise,
		"is_active":      p.IsActive,
		"is_deleted":     p.DeletedAt != nil,
		"image_url":      p.ImageURL,
		"thumbnail_url":  p.ThumbnailURL,
		"catalog_seq":    p.SyncSeq,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/zippyra_products/_doc/%s", strings.TrimRight(e.esEndpoint, "/"), p.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (e *ESSearchEngine) DeleteProductIndex(ctx context.Context, storeID, productID string) error {
	if e.esEndpoint == "" {
		return nil
	}

	url := fmt.Sprintf("%s/zippyra_products/_doc/%s", strings.TrimRight(e.esEndpoint, "/"), productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

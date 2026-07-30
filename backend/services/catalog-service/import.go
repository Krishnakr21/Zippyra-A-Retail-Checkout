package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/validator"
)

type ImportWorker struct {
	repo          Repository
	cacheMgr      CacheManager
	searchEngine  SearchEngine
	kafkaProducer *kafka.Producer
}

func NewImportWorker(repo Repository, cacheMgr CacheManager, searchEngine SearchEngine, producer *kafka.Producer) *ImportWorker {
	return &ImportWorker{
		repo:          repo,
		cacheMgr:      cacheMgr,
		searchEngine:  searchEngine,
		kafkaProducer: producer,
	}
}

func (w *ImportWorker) ProcessCSVImportJob(ctx context.Context, jobID string, csvReader io.Reader) {
	job, err := w.repo.GetImportJob(ctx, jobID)
	if err != nil || job == nil {
		logger.Error("Import job %s not found", jobID)
		return
	}

	job.Status = "PROCESSING"
	_ = w.repo.UpdateImportJob(ctx, job)

	reader := csv.NewReader(csvReader)
	header, err := reader.Read()
	if err != nil {
		job.Status = "FAILED"
		now := time.Now()
		job.CompletedAt = &now
		job.ErrorRows = append(job.ErrorRows, &ImportRowError{RowNumber: 0, Reason: "Invalid CSV header"})
		_ = w.repo.UpdateImportJob(ctx, job)
		return
	}

	colMap := make(map[string]int)
	for i, colName := range header {
		colMap[strings.TrimSpace(strings.ToLower(colName))] = i
	}

	var rowNumber int
	var createdCount int
	var maxSyncSeq int64

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Debug("CSV read error at row %d: %v", rowNumber+1, err)
			continue
		}
		rowNumber++

		barcode := getCSVCol(record, colMap, "barcode")
		name := getCSVCol(record, colMap, "name")
		desc := getCSVCol(record, colMap, "description")
		catID := getCSVCol(record, colMap, "category_id")
		if catID == "" {
			catID = getCSVCol(record, colMap, "category_name")
		}
		priceStr := getCSVCol(record, colMap, "price_paise")
		mrpStr := getCSVCol(record, colMap, "mrp_paise")
		hsn := getCSVCol(record, colMap, "hsn_code")
		imgURL := getCSVCol(record, colMap, "image_url")
		returnableStr := getCSVCol(record, colMap, "is_returnable")

		// 1. Validate Barcode Checksum
		if !validator.ValidateBarcode(barcode) {
			job.ErrorRows = append(job.ErrorRows, &ImportRowError{
				RowNumber: rowNumber,
				Barcode:   barcode,
				Reason:    "Invalid EAN-13 / UPC-A barcode checksum",
			})
			continue
		}

		// 2. Validate HSN Code presence
		hsnRate, err := w.repo.GetHSNRate(ctx, hsn)
		if err != nil || hsnRate == nil {
			job.ErrorRows = append(job.ErrorRows, &ImportRowError{
				RowNumber: rowNumber,
				Barcode:   barcode,
				Reason:    fmt.Sprintf("Unknown HSN code: %s", hsn),
			})
			continue
		}

		// 3. Validate Price > 0
		pricePaise, errPrice := strconv.ParseInt(priceStr, 10, 64)
		if errPrice != nil || pricePaise <= 0 {
			job.ErrorRows = append(job.ErrorRows, &ImportRowError{
				RowNumber: rowNumber,
				Barcode:   barcode,
				Reason:    "Price must be a positive integer in paise",
			})
			continue
		}

		mrpPaise, _ := strconv.ParseInt(mrpStr, 10, 64)
		if mrpPaise < pricePaise {
			mrpPaise = pricePaise
		}

		if name == "" {
			job.ErrorRows = append(job.ErrorRows, &ImportRowError{
				RowNumber: rowNumber,
				Barcode:   barcode,
				Reason:    "Product name cannot be empty",
			})
			continue
		}

		isReturnable := true
		if strings.ToLower(returnableStr) == "false" || returnableStr == "0" {
			isReturnable = false
		}

		var categoryPtr *string
		if catID != "" {
			categoryPtr = &catID
		}

		// Upsert / Create Product
		product := &Product{
			StoreID:        job.StoreID,
			ChainID:        job.ChainID,
			Barcode:        barcode,
			Name:           name,
			Description:    desc,
			CategoryID:     categoryPtr,
			PricePaise:     pricePaise,
			MRPPaise:       mrpPaise,
			HSNCode:        hsn,
			GSTRatePercent: hsnRate.GSTRatePercent,
			ImageURL:       imgURL,
			IsActive:       true,
			IsReturnable:   isReturnable,
		}

		// Check existing
		existing, _ := w.repo.GetProductByBarcode(ctx, job.StoreID, barcode)
		if existing != nil {
			product.ID = existing.ID
			_ = w.repo.UpdateProduct(ctx, product)
		} else {
			_ = w.repo.CreateProduct(ctx, product)
		}

		// Write-through Redis SKU cache update
		_ = w.cacheMgr.SetSKU(ctx, job.StoreID, barcode, product)

		// Async ES Indexing
		if w.searchEngine != nil {
			_ = w.searchEngine.IndexProduct(ctx, product)
		}

		if product.SyncSeq > maxSyncSeq {
			maxSyncSeq = product.SyncSeq
		}
		createdCount++
	}

	job.Status = "COMPLETED"
	job.TotalRows = rowNumber
	job.ProcessedRows = createdCount
	now := time.Now()
	job.CompletedAt = &now
	_ = w.repo.UpdateImportJob(ctx, job)

	// Publish ONE Kafka event for the entire completed import job
	if w.kafkaProducer != nil && createdCount > 0 {
		payload := map[string]interface{}{
			"store_id":      job.StoreID,
			"chain_id":      job.ChainID,
			"sync_seq":      maxSyncSeq,
			"changed_count": createdCount,
			"import_job_id": job.ID,
			"ts":            time.Now().Unix(),
		}
		_ = w.kafkaProducer.PublishEvent(ctx, "catalog.updated", job.StoreID, payload)
	}

	logger.Info("CSV import job %s completed: %d total rows, %d processed successfully, %d errors",
		job.ID, job.TotalRows, job.ProcessedRows, len(job.ErrorRows))
}

func getCSVCol(record []string, colMap map[string]int, colName string) string {
	idx, ok := colMap[colName]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

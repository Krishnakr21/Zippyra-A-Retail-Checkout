package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

type Repository interface {
	GetProductByBarcode(ctx context.Context, storeID, barcode string) (*Product, error)
	GetProductByID(ctx context.Context, id string) (*Product, error)
	CreateProduct(ctx context.Context, p *Product) error
	UpdateProduct(ctx context.Context, p *Product) error
	SoftDeleteProduct(ctx context.Context, id string) (*Product, error)
	SearchProductsPostgres(ctx context.Context, storeID, escapedQuery, categoryID string, page, pageSize int) ([]*Product, int64, error)
	GetDeltaSyncProducts(ctx context.Context, storeID string, sinceSeq int64, limit int) ([]*Product, []string, int64, bool, error)
	GetCategoriesByChain(ctx context.Context, chainID string) ([]*Category, error)
	GetHSNRate(ctx context.Context, hsnCode string) (*HsnGstRate, error)
	AdminListProducts(ctx context.Context, storeID string, page, pageSize int) ([]*Product, int64, error)
	CreateImportJob(ctx context.Context, job *CatalogImportJob) error
	GetImportJob(ctx context.Context, jobID string) (*CatalogImportJob, error)
	UpdateImportJob(ctx context.Context, job *CatalogImportJob) error
	CheckStoreHSN(ctx context.Context, storeID string) (int, []string, bool, error)
	UpdateProductImageStatus(ctx context.Context, s3RawKey, fullURL, thumbnailURL, status string) (*Product, error)
}

type PostgresRepository struct {
	database *db.DB
}

func NewPostgresRepository(database *db.DB) Repository {
	return &PostgresRepository{database: database}
}

func (p *PostgresRepository) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*Product, error) {
	query := `
		SELECT pr.id, pr.store_id, pr.chain_id, pr.barcode, pr.name, COALESCE(pr.description, ''),
		       pr.category_id, pr.price_paise, pr.mrp_paise, pr.hsn_code,
		       COALESCE(hsn.gst_rate_percent, 0), pr.is_active, pr.is_returnable,
		       COALESCE(pr.image_url, ''), COALESCE(pr.thumbnail_url, ''), COALESCE(pr.image_processing_status, 'PROCESSED'), pr.sync_seq, pr.deleted_at,
		       pr.created_at, pr.updated_at
		FROM products pr
		LEFT JOIN hsn_gst_rates hsn ON pr.hsn_code = hsn.hsn_code
		WHERE pr.store_id = $1 AND pr.barcode = $2 AND pr.deleted_at IS NULL AND pr.is_active = true
	`
	var pr Product
	err := p.database.QueryRowContext(ctx, query, storeID, barcode).Scan(
		&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description,
		&pr.CategoryID, &pr.PricePaise, &pr.MRPPaise, &pr.HSNCode,
		&pr.GSTRatePercent, &pr.IsActive, &pr.IsReturnable,
		&pr.ImageURL, &pr.ThumbnailURL, &pr.ImageProcessingStatus, &pr.SyncSeq, &pr.DeletedAt,
		&pr.CreatedAt, &pr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pr, err
}

func (p *PostgresRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	query := `
		SELECT pr.id, pr.store_id, pr.chain_id, pr.barcode, pr.name, COALESCE(pr.description, ''),
		       pr.category_id, pr.price_paise, pr.mrp_paise, pr.hsn_code,
		       COALESCE(hsn.gst_rate_percent, 0), pr.is_active, pr.is_returnable,
		       COALESCE(pr.image_url, ''), COALESCE(pr.thumbnail_url, ''), COALESCE(pr.image_processing_status, 'PROCESSED'), pr.sync_seq, pr.deleted_at,
		       pr.created_at, pr.updated_at
		FROM products pr
		LEFT JOIN hsn_gst_rates hsn ON pr.hsn_code = hsn.hsn_code
		WHERE pr.id = $1
	`
	var pr Product
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description,
		&pr.CategoryID, &pr.PricePaise, &pr.MRPPaise, &pr.HSNCode,
		&pr.GSTRatePercent, &pr.IsActive, &pr.IsReturnable,
		&pr.ImageURL, &pr.ThumbnailURL, &pr.ImageProcessingStatus, &pr.SyncSeq, &pr.DeletedAt,
		&pr.CreatedAt, &pr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pr, err
}

func (p *PostgresRepository) CreateProduct(ctx context.Context, pr *Product) error {
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}
	if pr.ImageProcessingStatus == "" {
		pr.ImageProcessingStatus = "PROCESSED"
	}
	query := `
		INSERT INTO products (id, store_id, chain_id, barcode, name, description, category_id,
		                    price_paise, mrp_paise, hsn_code, is_active, is_returnable, image_url, thumbnail_url, image_processing_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING sync_seq, created_at, updated_at
	`
	return p.database.QueryRowContext(ctx, query,
		pr.ID, pr.StoreID, pr.ChainID, pr.Barcode, pr.Name, pr.Description, pr.CategoryID,
		pr.PricePaise, pr.MRPPaise, pr.HSNCode, pr.IsActive, pr.IsReturnable, pr.ImageURL, pr.ThumbnailURL, pr.ImageProcessingStatus,
	).Scan(&pr.SyncSeq, &pr.CreatedAt, &pr.UpdatedAt)
}

func (p *PostgresRepository) UpdateProduct(ctx context.Context, pr *Product) error {
	if pr.ImageProcessingStatus == "" {
		pr.ImageProcessingStatus = "PROCESSED"
	}
	query := `
		UPDATE products
		SET name = $1, description = $2, category_id = $3, price_paise = $4, mrp_paise = $5,
		    hsn_code = $6, is_active = $7, is_returnable = $8, image_url = $9, thumbnail_url = $10,
		    image_processing_status = $11,
		    sync_seq = nextval('catalog_sync_seq'), updated_at = CURRENT_TIMESTAMP
		WHERE id = $12 AND deleted_at IS NULL
		RETURNING sync_seq, updated_at
	`
	return p.database.QueryRowContext(ctx, query,
		pr.Name, pr.Description, pr.CategoryID, pr.PricePaise, pr.MRPPaise,
		pr.HSNCode, pr.IsActive, pr.IsReturnable, pr.ImageURL, pr.ThumbnailURL, pr.ImageProcessingStatus, pr.ID,
	).Scan(&pr.SyncSeq, &pr.UpdatedAt)
}

func (p *PostgresRepository) UpdateProductImageStatus(ctx context.Context, s3RawKey, fullURL, thumbnailURL, status string) (*Product, error) {
	query := `
		UPDATE products
		SET image_processing_status = $1,
		    image_url = CASE WHEN $2 <> '' THEN $2 ELSE image_url END,
		    thumbnail_url = CASE WHEN $3 <> '' THEN $3 ELSE thumbnail_url END,
		    sync_seq = nextval('catalog_sync_seq'),
		    updated_at = CURRENT_TIMESTAMP
		WHERE (image_url LIKE '%' || $4 || '%' OR thumbnail_url LIKE '%' || $4 || '%' OR image_url = $4)
		  AND deleted_at IS NULL
		RETURNING id, store_id, chain_id, barcode, name, COALESCE(description, ''), category_id,
		          price_paise, mrp_paise, hsn_code, is_active, is_returnable, COALESCE(image_url, ''),
		          COALESCE(thumbnail_url, ''), COALESCE(image_processing_status, 'PROCESSED'), sync_seq, created_at, updated_at
	`
	var pr Product
	err := p.database.QueryRowContext(ctx, query, status, fullURL, thumbnailURL, s3RawKey).Scan(
		&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description, &pr.CategoryID,
		&pr.PricePaise, &pr.MRPPaise, &pr.HSNCode, &pr.IsActive, &pr.IsReturnable, &pr.ImageURL,
		&pr.ThumbnailURL, &pr.ImageProcessingStatus, &pr.SyncSeq, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pr, err
}

func (p *PostgresRepository) SoftDeleteProduct(ctx context.Context, id string) (*Product, error) {
	query := `
		UPDATE products
		SET deleted_at = CURRENT_TIMESTAMP, is_active = false, sync_seq = nextval('catalog_sync_seq'), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, store_id, chain_id, barcode, name, COALESCE(description, ''), category_id, price_paise, mrp_paise, hsn_code, is_active, is_returnable, sync_seq, deleted_at
	`
	var pr Product
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description, &pr.CategoryID,
		&pr.PricePaise, &pr.MRPPaise, &pr.HSNCode, &pr.IsActive, &pr.IsReturnable, &pr.SyncSeq, &pr.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pr, err
}

func (p *PostgresRepository) SearchProductsPostgres(ctx context.Context, storeID, escapedQuery, categoryID string, page, pageSize int) ([]*Product, int64, error) {
	offset := (page - 1) * pageSize

	baseWhere := `WHERE store_id = $1 AND deleted_at IS NULL AND is_active = true`
	args := []interface{}{storeID}
	argCount := 1

	if escapedQuery != "" {
		argCount++
		baseWhere += fmt.Sprintf(` AND name ILIKE $%d ESCAPE '\'`, argCount)
		args = append(args, "%"+escapedQuery+"%")
	}

	if categoryID != "" {
		argCount++
		baseWhere += fmt.Sprintf(` AND category_id = $%d`, argCount)
		args = append(args, categoryID)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products %s`, baseWhere)
	var total int64
	if err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT pr.id, pr.store_id, pr.chain_id, pr.barcode, pr.name, COALESCE(pr.description, ''),
		       pr.category_id, pr.price_paise, pr.mrp_paise, pr.hsn_code,
		       COALESCE(hsn.gst_rate_percent, 0), pr.is_active, pr.is_returnable,
		       COALESCE(pr.image_url, ''), COALESCE(pr.thumbnail_url, ''), pr.sync_seq, pr.created_at, pr.updated_at
		FROM products pr
		LEFT JOIN hsn_gst_rates hsn ON pr.hsn_code = hsn.hsn_code
		%s
		ORDER BY pr.name ASC
		LIMIT $%d OFFSET $%d
	`, baseWhere, argCount+1, argCount+2)

	queryArgs := append(args, pageSize, offset)
	rows, err := p.database.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var pr Product
		err := rows.Scan(
			&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description,
			&pr.CategoryID, &pr.PricePaise, &pr.MRPPaise, &pr.HSNCode,
			&pr.GSTRatePercent, &pr.IsActive, &pr.IsReturnable,
			&pr.ImageURL, &pr.ThumbnailURL, &pr.SyncSeq, &pr.CreatedAt, &pr.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, &pr)
	}

	return products, total, nil
}

func (p *PostgresRepository) GetDeltaSyncProducts(ctx context.Context, storeID string, sinceSeq int64, limit int) ([]*Product, []string, int64, bool, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}

	query := `
		SELECT id, store_id, chain_id, barcode, name, COALESCE(description, ''), category_id,
		       price_paise, mrp_paise, hsn_code, is_active, is_returnable,
		       COALESCE(image_url, ''), COALESCE(thumbnail_url, ''), sync_seq, deleted_at
		FROM products
		WHERE store_id = $1 AND sync_seq > $2
		ORDER BY sync_seq ASC
		LIMIT $3
	`

	rows, err := p.database.QueryContext(ctx, query, storeID, sinceSeq, limit+1)
	if err != nil {
		return nil, nil, sinceSeq, false, err
	}
	defer rows.Close()

	var products []*Product
	var deletedIDs []string
	var maxSeq = sinceSeq
	count := 0
	hasMore := false

	for rows.Next() {
		count++
		if count > limit {
			hasMore = true
			break
		}

		var pr Product
		err := rows.Scan(
			&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description, &pr.CategoryID,
			&pr.PricePaise, &pr.MRPPaise, &pr.HSNCode, &pr.IsActive, &pr.IsReturnable,
			&pr.ImageURL, &pr.ThumbnailURL, &pr.SyncSeq, &pr.DeletedAt,
		)
		if err != nil {
			return nil, nil, sinceSeq, false, err
		}

		if pr.SyncSeq > maxSeq {
			maxSeq = pr.SyncSeq
		}

		if pr.DeletedAt != nil {
			deletedIDs = append(deletedIDs, pr.ID)
		} else {
			products = append(products, &pr)
		}
	}

	return products, deletedIDs, maxSeq, hasMore, nil
}

func (p *PostgresRepository) GetCategoriesByChain(ctx context.Context, chainID string) ([]*Category, error) {
	query := `
		SELECT id, chain_id, name, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE chain_id = $1 AND is_active = true
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := p.database.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	catMap := make(map[string]*Category)

	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ChainID, &c.Name, &c.ParentID, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Children = []*Category{}
		categories = append(categories, &c)
		catMap[c.ID] = &c
	}

	// Build nested category tree
	var tree []*Category
	for _, c := range categories {
		if c.ParentID == nil || *c.ParentID == "" {
			tree = append(tree, c)
		} else if parent, ok := catMap[*c.ParentID]; ok {
			parent.Children = append(parent.Children, c)
		} else {
			tree = append(tree, c)
		}
	}

	return tree, nil
}

func (p *PostgresRepository) GetHSNRate(ctx context.Context, hsnCode string) (*HsnGstRate, error) {
	query := `SELECT hsn_code, gst_rate_percent, cess_percent, COALESCE(description, '') FROM hsn_gst_rates WHERE hsn_code = $1`
	var rate HsnGstRate
	err := p.database.QueryRowContext(ctx, query, hsnCode).Scan(&rate.HSNCode, &rate.GSTRatePercent, &rate.CessPercent, &rate.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rate, err
}

func (p *PostgresRepository) CreateImportJob(ctx context.Context, job *CatalogImportJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	errJSON, _ := json.Marshal(job.ErrorRows)
	query := `INSERT INTO catalog_import_jobs (id, store_id, chain_id, status, total_rows, processed_rows, error_rows) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := p.database.ExecContext(ctx, query, job.ID, job.StoreID, job.ChainID, job.Status, job.TotalRows, job.ProcessedRows, errJSON)
	return err
}

func (p *PostgresRepository) GetImportJob(ctx context.Context, jobID string) (*CatalogImportJob, error) {
	query := `SELECT id, store_id, chain_id, status, total_rows, processed_rows, error_rows, created_at, completed_at FROM catalog_import_jobs WHERE id = $1`
	var job CatalogImportJob
	var errJSON []byte
	err := p.database.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.StoreID, &job.ChainID, &job.Status, &job.TotalRows, &job.ProcessedRows, &errJSON, &job.CreatedAt, &job.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(errJSON) > 0 {
		_ = json.Unmarshal(errJSON, &job.ErrorRows)
	}
	return &job, nil
}

func (p *PostgresRepository) UpdateImportJob(ctx context.Context, job *CatalogImportJob) error {
	errJSON, _ := json.Marshal(job.ErrorRows)
	query := `UPDATE catalog_import_jobs SET status = $1, total_rows = $2, processed_rows = $3, error_rows = $4, completed_at = $5 WHERE id = $6`
	_, err := p.database.ExecContext(ctx, query, job.Status, job.TotalRows, job.ProcessedRows, errJSON, job.CompletedAt, job.ID)
	return err
}

// MemoryRepository for testing
type MemoryRepository struct {
	mu         sync.RWMutex
	products   map[string]*Product
	hsnRates   map[string]*HsnGstRate
	categories map[string]*Category
	importJobs map[string]*CatalogImportJob
	seqCounter int64
}

func NewMemoryRepository() Repository {
	m := &MemoryRepository{
		products:   make(map[string]*Product),
		hsnRates:   make(map[string]*HsnGstRate),
		categories: make(map[string]*Category),
		importJobs: make(map[string]*CatalogImportJob),
		seqCounter: 1,
	}

	// Seed standard HSN tax rates for tests
	m.hsnRates["0405"] = &HsnGstRate{HSNCode: "0405", GSTRatePercent: 12.0, Description: "Butter"}
	m.hsnRates["0901"] = &HsnGstRate{HSNCode: "0901", GSTRatePercent: 5.0, Description: "Coffee"}
	m.hsnRates["1905"] = &HsnGstRate{HSNCode: "1905", GSTRatePercent: 18.0, Description: "Biscuits"}

	// Seed categories
	m.categories["cat-1"] = &Category{ID: "cat-1", ChainID: "chain-hq-001", Name: "Grocery", SortOrder: 1, IsActive: true}
	m.categories["cat-2"] = &Category{ID: "cat-2", ChainID: "chain-hq-001", Name: "Beverages", SortOrder: 2, IsActive: true}
	m.categories["cat-3"] = &Category{ID: "cat-3", ChainID: "chain-hq-001", Name: "Snacks", SortOrder: 3, IsActive: true}
	m.categories["cat-4"] = &Category{ID: "cat-4", ChainID: "chain-hq-001", Name: "Dairy", SortOrder: 4, IsActive: true}
	m.categories["cat-5"] = &Category{ID: "cat-5", ChainID: "chain-hq-001", Name: "Personal Care", SortOrder: 5, IsActive: true}
	m.categories["cat-6"] = &Category{ID: "cat-6", ChainID: "chain-hq-001", Name: "Apparel", SortOrder: 6, IsActive: true}

	// Seed products
	catDairy := "cat-4"
	catSnacks := "cat-3"
	catBev := "cat-2"
	catGroc := "cat-1"

	pList := []*Product{
		{
			ID:         "p-1",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901262010051",
			Name:       "Amul Butter 500g",
			CategoryID: &catDairy,
			PricePaise: 28000,
			MRPPaise:   31000,
			HSNCode:    "0405",
			IsActive:   true,
		},
		{
			ID:         "p-2",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901491100012",
			Name:       "Lay's Classic 150g",
			CategoryID: &catSnacks,
			PricePaise: 4500,
			MRPPaise:   4500,
			HSNCode:    "1905",
			IsActive:   true,
		},
		{
			ID:         "p-3",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901052000018",
			Name:       "Tata Tea Gold 250g",
			CategoryID: &catGroc,
			PricePaise: 19500,
			MRPPaise:   22000,
			HSNCode:    "0901",
			IsActive:   true,
		},
		{
			ID:         "p-4",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901063000025",
			Name:       "Britannia Bread 400g",
			CategoryID: &catGroc,
			PricePaise: 4500,
			MRPPaise:   4500,
			HSNCode:    "1905",
			IsActive:   true,
		},
		{
			ID:         "p-5",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901262020012",
			Name:       "Amul Taza Milk 1L",
			CategoryID: &catDairy,
			PricePaise: 6200,
			MRPPaise:   6200,
			HSNCode:    "0405",
			IsActive:   true,
		},
		{
			ID:         "p-6",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "8901000000064",
			Name:       "Eggs 6-pack",
			CategoryID: &catDairy,
			PricePaise: 7800,
			MRPPaise:   7800,
			HSNCode:    "0405",
			IsActive:   true,
		},
		{
			ID:         "p-7",
			StoreID:    "store-1",
			ChainID:    "chain-hq-001",
			Barcode:    "9002490100017",
			Name:       "Red Bull Energy 250ml",
			CategoryID: &catBev,
			PricePaise: 7500,
			MRPPaise:   15000,
			HSNCode:    "0901",
			IsActive:   true,
		},
	}

	for _, p := range pList {
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		m.products[p.ID] = p
	}

	// Seed products from demo_catalog.csv if available
	csvPaths := []string{
		"demo_catalog.csv",
		"../../demo_catalog.csv",
		"../demo_catalog.csv",
		"/Users/krishna/Downloads/Fatima/Zippyra/demo_catalog.csv",
	}

	for _, path := range csvPaths {
		file, err := os.Open(path)
		if err == nil {
			r := csv.NewReader(file)
			header, err := r.Read()
			if err == nil {
				colMap := make(map[string]int)
				for i, colName := range header {
					colMap[strings.TrimSpace(strings.ToLower(colName))] = i
				}
				rowIdx := 1
				for {
					rec, err := r.Read()
					if err == io.EOF {
						break
					}
					if err != nil {
						continue
					}
					barcode := getCSVCol(rec, colMap, "barcode")
					name := getCSVCol(rec, colMap, "name")
					priceStr := getCSVCol(rec, colMap, "price_paise")
					mrpStr := getCSVCol(rec, colMap, "mrp_paise")
					hsn := getCSVCol(rec, colMap, "hsn_code")
					catName := getCSVCol(rec, colMap, "category_name")
					if catName == "" {
						catName = getCSVCol(rec, colMap, "category_id")
					}
					imgURL := getCSVCol(rec, colMap, "image_url")
					retStr := getCSVCol(rec, colMap, "is_returnable")

					if barcode == "" || name == "" {
						continue
					}

					priceP, _ := strconv.ParseInt(priceStr, 10, 64)
					mrpP, _ := strconv.ParseInt(mrpStr, 10, 64)
					if priceP <= 0 {
						priceP = 5000
					}
					if mrpP < priceP {
						mrpP = priceP
					}

					catID := strings.ToLower(strings.ReplaceAll(catName, " ", "-"))
					if catID != "" && m.categories[catID] == nil {
						m.categories[catID] = &Category{
							ID:        catID,
							ChainID:   "chain-hq-001",
							Name:      strings.Title(catName),
							SortOrder: len(m.categories) + 1,
							IsActive:  true,
						}
					}

					pID := fmt.Sprintf("demo-%d", rowIdx)
					rowIdx++
					m.seqCounter++
					isRet := strings.ToLower(retStr) != "false"

					var catPtr *string
					if catID != "" {
						catPtr = &catID
					}

					prod := &Product{
						ID:             pID,
						StoreID:        "store-1",
						ChainID:        "chain-hq-001",
						Barcode:        barcode,
						Name:           name,
						CategoryID:     catPtr,
						PricePaise:     priceP,
						MRPPaise:       mrpP,
						HSNCode:        hsn,
						GSTRatePercent: 12.0,
						ImageURL:       imgURL,
						IsActive:       true,
						IsReturnable:   isRet,
						SyncSeq:        m.seqCounter,
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
					}
					m.products[pID] = prod
				}
			}
			file.Close()
			break
		}
	}

	return m
}

func (m *MemoryRepository) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.products {
		if p.StoreID == storeID && p.Barcode == barcode && p.DeletedAt == nil && p.IsActive {
			if rate, ok := m.hsnRates[p.HSNCode]; ok {
				p.GSTRatePercent = rate.GSTRatePercent
			}
			return p, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.products[id]
	if !ok {
		return nil, nil
	}
	if rate, ok := m.hsnRates[p.HSNCode]; ok {
		p.GSTRatePercent = rate.GSTRatePercent
	}
	return p, nil
}

func (m *MemoryRepository) CreateProduct(ctx context.Context, p *Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	m.seqCounter++
	p.SyncSeq = m.seqCounter
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	m.products[p.ID] = p
	return nil
}

func (m *MemoryRepository) UpdateProduct(ctx context.Context, p *Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.products[p.ID]
	if !ok || existing.DeletedAt != nil {
		return errors.New("product not found")
	}

	m.seqCounter++
	p.SyncSeq = m.seqCounter
	p.UpdatedAt = time.Now()
	m.products[p.ID] = p
	return nil
}

func (m *MemoryRepository) SoftDeleteProduct(ctx context.Context, id string) (*Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.products[id]
	if !ok || p.DeletedAt != nil {
		return nil, nil
	}

	now := time.Now()
	p.DeletedAt = &now
	p.IsActive = false
	m.seqCounter++
	p.SyncSeq = m.seqCounter
	p.UpdatedAt = now
	return p, nil
}

func (m *MemoryRepository) SearchProductsPostgres(ctx context.Context, storeID, escapedQuery, categoryID string, page, pageSize int) ([]*Product, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cleanQuery := strings.ReplaceAll(escapedQuery, "\\%", "%")
	cleanQuery = strings.ReplaceAll(cleanQuery, "\\_", "_")
	cleanQuery = strings.ToLower(cleanQuery)

	var matched []*Product
	for _, p := range m.products {
		if p.StoreID != storeID || p.DeletedAt != nil || !p.IsActive {
			continue
		}
		if categoryID != "" && (p.CategoryID == nil || *p.CategoryID != categoryID) {
			continue
		}
		if cleanQuery != "" && !strings.Contains(strings.ToLower(p.Name), cleanQuery) {
			continue
		}
		if rate, ok := m.hsnRates[p.HSNCode]; ok {
			p.GSTRatePercent = rate.GSTRatePercent
		}
		matched = append(matched, p)
	}

	total := int64(len(matched))
	offset := (page - 1) * pageSize
	if offset >= len(matched) {
		return []*Product{}, total, nil
	}

	end := offset + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (m *MemoryRepository) GetDeltaSyncProducts(ctx context.Context, storeID string, sinceSeq int64, limit int) ([]*Product, []string, int64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var products []*Product
	var deletedIDs []string
	var maxSeq = sinceSeq

	for _, p := range m.products {
		if p.StoreID == storeID && p.SyncSeq > sinceSeq {
			if p.SyncSeq > maxSeq {
				maxSeq = p.SyncSeq
			}
			if p.DeletedAt != nil {
				deletedIDs = append(deletedIDs, p.ID)
			} else {
				products = append(products, p)
			}
		}
	}

	hasMore := len(products)+len(deletedIDs) > limit
	if limit > 0 && len(products) > limit {
		products = products[:limit]
	}

	return products, deletedIDs, maxSeq, hasMore, nil
}

func (m *MemoryRepository) GetCategoriesByChain(ctx context.Context, chainID string) ([]*Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Category
	for _, c := range m.categories {
		if c.ChainID == chainID && c.IsActive {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MemoryRepository) GetHSNRate(ctx context.Context, hsnCode string) (*HsnGstRate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rate, ok := m.hsnRates[hsnCode]
	if !ok {
		return nil, nil
	}
	return rate, nil
}

func (m *MemoryRepository) CreateHSNRate(ctx context.Context, rate *HsnGstRate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hsnRates[rate.HSNCode] = rate
}

func (m *MemoryRepository) CreateImportJob(ctx context.Context, job *CatalogImportJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	job.CreatedAt = time.Now()
	m.importJobs[job.ID] = job
	return nil
}

func (m *MemoryRepository) GetImportJob(ctx context.Context, jobID string) (*CatalogImportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.importJobs[jobID]
	if !ok {
		return nil, nil
	}
	return job, nil
}

func (m *MemoryRepository) UpdateImportJob(ctx context.Context, job *CatalogImportJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.importJobs[job.ID] = job
	return nil
}

func (p *PostgresRepository) AdminListProducts(ctx context.Context, storeID string, page, pageSize int) ([]*Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	countQuery := `SELECT COUNT(*) FROM products WHERE store_id = $1 AND deleted_at IS NULL`
	if err := p.database.QueryRowContext(ctx, countQuery, storeID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	query := `
		SELECT pr.id, pr.store_id, pr.chain_id, pr.barcode, pr.name, COALESCE(pr.description, ''),
		       pr.category_id, pr.price_paise, pr.mrp_paise, pr.hsn_code,
		       COALESCE(hsn.gst_rate_percent, 0), pr.is_active, pr.is_returnable,
		       COALESCE(pr.image_url, ''), COALESCE(pr.thumbnail_url, ''), pr.sync_seq, pr.deleted_at,
		       pr.created_at, pr.updated_at
		FROM products pr
		LEFT JOIN hsn_gst_rates hsn ON pr.hsn_code = hsn.hsn_code
		WHERE pr.store_id = $1 AND pr.deleted_at IS NULL
		ORDER BY pr.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.database.QueryContext(ctx, query, storeID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var pr Product
		if err := rows.Scan(
			&pr.ID, &pr.StoreID, &pr.ChainID, &pr.Barcode, &pr.Name, &pr.Description,
			&pr.CategoryID, &pr.PricePaise, &pr.MRPPaise, &pr.HSNCode,
			&pr.GSTRatePercent, &pr.IsActive, &pr.IsReturnable,
			&pr.ImageURL, &pr.ThumbnailURL, &pr.SyncSeq, &pr.DeletedAt,
			&pr.CreatedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &pr)
	}

	if products == nil {
		products = []*Product{}
	}

	return products, total, nil
}

func (m *MemoryRepository) AdminListProducts(ctx context.Context, storeID string, page, pageSize int) ([]*Product, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*Product
	for _, p := range m.products {
		if p.StoreID == storeID && p.DeletedAt == nil {
			matched = append(matched, p)
		}
	}

	total := int64(len(matched))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= len(matched) {
		return []*Product{}, total, nil
	}

	end := offset + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (p *PostgresRepository) CheckStoreHSN(ctx context.Context, storeID string) (int, []string, bool, error) {
	query := `
		SELECT DISTINCT pr.hsn_code, (hsn.hsn_code IS NOT NULL) as is_mapped
		FROM products pr
		LEFT JOIN hsn_gst_rates hsn ON pr.hsn_code = hsn.hsn_code
		WHERE pr.store_id = $1 AND pr.deleted_at IS NULL
	`
	rows, err := p.database.QueryContext(ctx, query, storeID)
	if err != nil {
		return 0, nil, false, err
	}
	defer rows.Close()

	totalHSN := 0
	var missingHSN []string
	for rows.Next() {
		var hsnCode string
		var isMapped bool
		if err := rows.Scan(&hsnCode, &isMapped); err != nil {
			return 0, nil, false, err
		}
		totalHSN++
		if !isMapped && hsnCode != "" {
			missingHSN = append(missingHSN, hsnCode)
		}
	}
	if missingHSN == nil {
		missingHSN = []string{}
	}
	isReady := len(missingHSN) == 0
	return totalHSN, missingHSN, isReady, nil
}

func (m *MemoryRepository) CheckStoreHSN(ctx context.Context, storeID string) (int, []string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hsnSet := make(map[string]bool)
	for _, p := range m.products {
		if p.StoreID == storeID && p.DeletedAt == nil && p.HSNCode != "" {
			hsnSet[p.HSNCode] = true
		}
	}

	totalHSN := len(hsnSet)
	var missingHSN []string
	for hsn := range hsnSet {
		if _, exists := m.hsnRates[hsn]; !exists {
			missingHSN = append(missingHSN, hsn)
		}
	}
	if missingHSN == nil {
		missingHSN = []string{}
	}
	isReady := len(missingHSN) == 0
	return totalHSN, missingHSN, isReady, nil
}

func (m *MemoryRepository) UpdateProductImageStatus(ctx context.Context, s3RawKey, fullURL, thumbnailURL, status string) (*Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.products {
		if p.DeletedAt == nil && (strings.Contains(p.ImageURL, s3RawKey) || strings.Contains(p.ThumbnailURL, s3RawKey) || p.ImageURL == s3RawKey) {
			p.ImageProcessingStatus = status
			if fullURL != "" {
				p.ImageURL = fullURL
			}
			if thumbnailURL != "" {
				p.ThumbnailURL = thumbnailURL
			}
			m.seqCounter++
			p.SyncSeq = m.seqCounter
			p.UpdatedAt = time.Now()
			return p, nil
		}
	}
	return nil, nil
}

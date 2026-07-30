package main

import (
	"net/http"

	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(customerHandler *CatalogHandler, adminHandler *AdminCatalogHandler, healthChecker func() bool) http.Handler {
	mux := http.NewServeMux()

	// Health routes
	mux.HandleFunc("/healthz/live", health.LiveHandler)
	mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if healthChecker != nil && !healthChecker() {
			http.Error(w, "NOT READY", http.StatusServiceUnavailable)
			return
		}
		health.ReadyHandler(w, r)
	})
	mux.HandleFunc("/healthz/startup", health.StartupHandler)

	// Customer Routes
	mux.HandleFunc("/v1/catalog/barcode/", customerHandler.HandleGetByBarcode)
	mux.HandleFunc("/v1/catalog/search", customerHandler.HandleSearch)
	mux.HandleFunc("/v1/catalog/categories", customerHandler.HandleGetCategories)
	mux.HandleFunc("/v1/catalog/sync", customerHandler.HandleDeltaSync)

	// Admin Product Routes
	mux.HandleFunc("/v1/catalog/admin/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			customerHandler.HandleAdminListProducts(w, r)
		} else if r.Method == http.MethodPost {
			adminHandler.HandleCreateProduct(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/catalog/admin/products/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			adminHandler.HandleUpdateProduct(w, r)
		} else if r.Method == http.MethodDelete {
			adminHandler.HandleDeleteProduct(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin Import Routes
	mux.HandleFunc("/v1/catalog/admin/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			adminHandler.HandleCSVImport(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/catalog/admin/import/", func(w http.ResponseWriter, r *http.Request) {
		adminHandler.HandleGetImportStatus(w, r)
	})

	mux.HandleFunc("/v1/catalog/admin/hsn-check", customerHandler.HandleAdminHSNCheck)
	mux.HandleFunc("/v1/catalog/admin/reindex", adminHandler.HandleReindexES)

	// Internal Webhook Routes
	mux.HandleFunc("/v1/catalog/internal/image-processed", adminHandler.HandleImageProcessedWebhook)

	// Standard middleware wrapper (1MB body limit for standard endpoints)
	handler := middleware.MaxBytesMiddleware(1048576)(mux)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	return handler
}

package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(
	customerHandler *CartHandler,
	internalHandler *InternalCartHandler,
	adminHandler *OfferAdminHandler,
	couponAdminHandler *CouponAdminHandler,
	healthChecker func() bool,
) http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)

	// Health routes
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if healthChecker != nil && !healthChecker() {
			http.Error(w, "NOT READY", http.StatusServiceUnavailable)
			return
		}
		health.ReadyHandler(w, r)
	}).Methods(http.MethodGet)
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods(http.MethodGet)

	// Admin Offer Authoring & Preview Routes
	if adminHandler != nil {
		r.HandleFunc("/v1/cart/admin/offers", adminHandler.CreateOfferHandler).Methods(http.MethodPost)
		r.HandleFunc("/v1/cart/admin/offers", adminHandler.ListOffersHandler).Methods(http.MethodGet)
		r.HandleFunc("/v1/cart/admin/offers/{id}", adminHandler.GetOfferHandler).Methods(http.MethodGet)
		r.HandleFunc("/v1/cart/admin/offers/{id}", adminHandler.UpdateOfferHandler).Methods(http.MethodPut)
		r.HandleFunc("/v1/cart/admin/offers/{id}", adminHandler.DeleteOfferHandler).Methods(http.MethodDelete)
		r.HandleFunc("/v1/cart/admin/offers/{id}/toggle", adminHandler.ToggleOfferHandler).Methods(http.MethodPost)
		r.HandleFunc("/v1/cart/admin/offers/{store_id}/preview", adminHandler.PreviewCompiledRulesHandler).Methods(http.MethodGet)
	}

	// Admin Coupon Authoring Routes
	if couponAdminHandler != nil {
		r.HandleFunc("/v1/cart/admin/coupons", couponAdminHandler.HandleCreateCoupon).Methods(http.MethodPost)
		r.HandleFunc("/v1/cart/admin/coupons", couponAdminHandler.HandleListCoupons).Methods(http.MethodGet)
		r.HandleFunc("/v1/cart/admin/coupons/{id}", couponAdminHandler.HandleUpdateCoupon).Methods(http.MethodPut)
		r.HandleFunc("/v1/cart/admin/coupons/{id}", couponAdminHandler.HandleDeleteCoupon).Methods(http.MethodDelete)
		r.HandleFunc("/v1/cart/admin/coupons/{id}/toggle", couponAdminHandler.HandleToggleCoupon).Methods(http.MethodPost)
	}

	// Customer Cart Routes
	r.HandleFunc("/v1/cart/scan", customerHandler.HandleScanItem).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/v1/cart/checkout/init", customerHandler.HandleCheckoutInit).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/v1/cart/item/{barcode}", customerHandler.HandleUpdateItem).Methods(http.MethodPut, http.MethodOptions)
	r.HandleFunc("/v1/cart/item/{barcode}", customerHandler.HandleDeleteItem).Methods(http.MethodDelete, http.MethodOptions)
	r.HandleFunc("/v1/cart/coupon/apply", customerHandler.HandleApplyCoupon).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/v1/cart/coupon", customerHandler.HandleRemoveCoupon).Methods(http.MethodDelete, http.MethodOptions)
	r.HandleFunc("/v1/cart", customerHandler.HandleGetCart).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/v1/cart", customerHandler.HandleClearCart).Methods(http.MethodDelete, http.MethodOptions)

	// Internal Service Routes
	r.HandleFunc("/v1/cart/internal/checkout-session/{id}", internalHandler.HandleGetCheckoutSession).Methods(http.MethodGet)
	r.HandleFunc("/v1/cart/internal/checkout-session/{id}", internalHandler.HandleConsumeCheckoutSession).Methods(http.MethodPost)

	// Standard middleware wrapper (1MB max body limit)
	var handler http.Handler = r
	handler = middleware.MaxBytesMiddleware(1048576)(handler)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	return handler
}

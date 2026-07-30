package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/featureflags"
	"github.com/zippyra/backend/shared/gst"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/validator"
)

type CartHandler struct {
	cartStore     CartStore
	holdManager   HoldManager
	offerEngine   OfferEngine
	checkoutRepo  CheckoutSessionRepository
	lockManager   LockManager
	catalogEngine CatalogLookupEngine
	eventProc     *EventProcessor
	offersRepo    OfferRepository
	jwtSecret     string
}

type CatalogLookupEngine interface {
	GetProductByBarcode(ctx context.Context, storeID, barcode string) (*ProductDTO, error)
}

type ProductDTO struct {
	ID             string  `json:"id"`
	StoreID        string  `json:"store_id"`
	Barcode        string  `json:"barcode"`
	Name           string  `json:"name"`
	PricePaise     int64   `json:"price_paise"`
	MRPPaise       int64   `json:"mrp_paise"`
	HSNCode        string  `json:"hsn_code"`
	GSTRatePercent float64 `json:"gst_rate_percent"`
	CategoryID     string  `json:"category_id,omitempty"`
}

type DefaultCatalogLookupEngine struct {
	catalogBaseURL string
	client         *http.Client
}

func NewDefaultCatalogLookupEngine(catalogBaseURL string) CatalogLookupEngine {
	if catalogBaseURL == "" {
		catalogBaseURL = "http://localhost:8084"
	}
	return &DefaultCatalogLookupEngine{
		catalogBaseURL: catalogBaseURL,
		client:         &http.Client{Timeout: 800 * time.Millisecond},
	}
}

func (c *DefaultCatalogLookupEngine) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*ProductDTO, error) {
	url := fmt.Sprintf("%s/v1/catalog/barcode/%s?store_id=%s", c.catalogBaseURL, barcode, storeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewAPIError(errors.CodeProductNotFound, "Product not found", nil)
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAPIError(errors.CodeInternalError, "Catalog service error", nil)
	}

	var p ProductDTO
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func NewCartHandler(
	cartStore CartStore,
	holdManager HoldManager,
	offerEngine OfferEngine,
	checkoutRepo CheckoutSessionRepository,
	lockManager LockManager,
	catalogEngine CatalogLookupEngine,
	eventProc *EventProcessor,
	jwtSecret string,
	offersRepo ...OfferRepository,
) *CartHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	var oRepo OfferRepository
	if len(offersRepo) > 0 {
		oRepo = offersRepo[0]
	}
	return &CartHandler{
		cartStore:     cartStore,
		holdManager:   holdManager,
		offerEngine:   offerEngine,
		checkoutRepo:  checkoutRepo,
		lockManager:   lockManager,
		catalogEngine: catalogEngine,
		eventProc:     eventProc,
		offersRepo:    oRepo,
		jwtSecret:     jwtSecret,
	}
}

// POST /v1/cart/scan (Target P99 < 100ms)
func (h *CartHandler) HandleScanItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	var req ScanItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Barcode == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid barcode or payload", nil)
		return
	}

	if req.Qty <= 0 {
		req.Qty = 1
	}

	// a. Checksum validation
	if !validator.ValidateBarcode(req.Barcode) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeBarcodeInvalid, "Invalid EAN-13 or UPC-A barcode checksum", nil)
		return
	}

	// b. Lock check
	lockAcquired, _ := h.lockManager.AcquireCheckoutLock(r.Context(), userID)
	if !lockAcquired {
		// If lock held, verify if already locked
		errors.WriteError(w, http.StatusConflict, errors.CodeCartLocked, "Cart is locked due to active checkout in progress", nil)
		return
	}
	// Release temporary lock after scan operation finishes
	defer func() { _ = h.lockManager.ReleaseCheckoutLock(r.Context(), userID) }()

	// c. Product lookup
	ctx, cancel := context.WithTimeout(r.Context(), 1000*time.Millisecond)
	defer cancel()

	product, err := h.catalogEngine.GetProductByBarcode(ctx, sessionStoreID, req.Barcode)
	if err != nil || product == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeProductNotFound, "Product not found", nil)
		return
	}

	// d. Reserve Hold
	if err := h.holdManager.CheckStockAndReserveHold(ctx, sessionStoreID, userID, req.Barcode, req.Qty); err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			errors.WriteError(w, http.StatusConflict, apiErr.Code, apiErr.Message, nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Inventory hold error", nil)
		return
	}

	// e. Fetch current cart & update
	currentItems, _, _ := h.cartStore.GetCart(ctx, sessionStoreID, userID)
	var existingItem *CartItem
	for _, item := range currentItems {
		if item.Barcode == req.Barcode {
			existingItem = item
			break
		}
	}

	newQty := req.Qty
	if existingItem != nil {
		newQty += existingItem.Qty
	}

	cartItem := &CartItem{
		Barcode:            req.Barcode,
		Name:               product.Name,
		Qty:                newQty,
		PricePaiseSnapshot: product.PricePaise,
		PricePaise:         product.PricePaise,
		LineTotalPaise:     product.PricePaise * int64(newQty),
		HSNCode:            product.HSNCode,
		CategoryID:         product.CategoryID,
	}

	if err := h.cartStore.UpsertCartItem(ctx, sessionStoreID, userID, cartItem); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update cart", nil)
		return
	}

	// Return updated summary
	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// GET /v1/cart
func (h *CartHandler) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// PUT /v1/cart/item/{barcode}
func (h *CartHandler) HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var barcode string
	for i, p := range parts {
		if p == "item" && i+1 < len(parts) {
			barcode = parts[i+1]
			break
		}
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.Qty <= 0 {
		h.deleteSingleItem(w, r, sessionStoreID, userID, barcode)
		return
	}

	currentItems, _, _ := h.cartStore.GetCart(r.Context(), sessionStoreID, userID)
	var existingItem *CartItem
	for _, item := range currentItems {
		if item.Barcode == barcode {
			existingItem = item
			break
		}
	}

	if existingItem == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeProductNotFound, "Item not found in cart", nil)
		return
	}

	delta := req.Qty - existingItem.Qty
	if delta > 0 {
		// Reserve extra hold
		if err := h.holdManager.CheckStockAndReserveHold(r.Context(), sessionStoreID, userID, barcode, delta); err != nil {
			if apiErr, ok := err.(*errors.APIError); ok {
				errors.WriteError(w, http.StatusConflict, apiErr.Code, apiErr.Message, nil)
				return
			}
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Inventory hold error", nil)
			return
		}
	} else if delta < 0 {
		// Release excess hold
		_ = h.holdManager.ReleaseHold(r.Context(), sessionStoreID, userID, barcode, -delta)
	}

	existingItem.Qty = req.Qty
	existingItem.LineTotalPaise = existingItem.PricePaiseSnapshot * int64(req.Qty)
	_ = h.cartStore.UpsertCartItem(r.Context(), sessionStoreID, userID, existingItem)

	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// DELETE /v1/cart/item/{barcode}
func (h *CartHandler) HandleDeleteItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var barcode string
	for i, p := range parts {
		if p == "item" && i+1 < len(parts) {
			barcode = parts[i+1]
			break
		}
	}

	h.deleteSingleItem(w, r, sessionStoreID, userID, barcode)
}

func (h *CartHandler) deleteSingleItem(w http.ResponseWriter, r *http.Request, storeID, userID, barcode string) {
	items, _, _ := h.cartStore.GetCart(r.Context(), storeID, userID)
	for _, item := range items {
		if item.Barcode == barcode {
			_ = h.holdManager.ReleaseHold(r.Context(), storeID, userID, barcode, item.Qty)
			break
		}
	}
	_ = h.cartStore.RemoveCartItem(r.Context(), storeID, userID, barcode)
	h.writeCartSummaryResponse(w, r, storeID, userID)
}

// DELETE /v1/cart
func (h *CartHandler) HandleClearCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	items, _, _ := h.cartStore.GetCart(r.Context(), sessionStoreID, userID)
	_ = h.holdManager.ReleaseAllUserHolds(r.Context(), sessionStoreID, userID, items)
	_ = h.cartStore.ClearCart(r.Context(), sessionStoreID, userID)

	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// POST /v1/cart/coupon/apply
func (h *CartHandler) HandleApplyCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	var req ApplyCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Code) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Coupon code is required", nil)
		return
	}

	code := strings.TrimSpace(strings.ToUpper(req.Code))

	if h.offersRepo != nil {
		chainID, _ := h.offersRepo.GetStoreChainID(r.Context(), sessionStoreID)
		coupon, err := h.offersRepo.GetCouponByCode(r.Context(), chainID, code)
		if err == nil && coupon != nil {
			if !coupon.IsActive {
				errors.WriteError(w, http.StatusBadRequest, "COUPON_INACTIVE", "Coupon is no longer active", nil)
				return
			}
			if coupon.MaxUses != nil && coupon.CurrentUseCount >= *coupon.MaxUses {
				errors.WriteError(w, http.StatusBadRequest, "COUPON_MAX_USES_EXCEEDED", "Coupon global maximum usage limit reached", nil)
				return
			}
			reds, _ := h.offersRepo.GetUserCouponRedemptions(r.Context(), coupon.ID, userID)
			if reds >= coupon.MaxUsesPerCustomer {
				errors.WriteError(w, http.StatusBadRequest, "COUPON_MAX_USES_EXCEEDED", "Coupon per-customer usage limit reached", nil)
				return
			}
		}
	}

	_ = h.cartStore.SetCoupon(r.Context(), sessionStoreID, userID, code)

	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// DELETE /v1/cart/coupon
func (h *CartHandler) HandleRemoveCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	_ = h.cartStore.RemoveCoupon(r.Context(), sessionStoreID, userID)
	h.writeCartSummaryResponse(w, r, sessionStoreID, userID)
}

// POST /v1/cart/checkout/init
func (h *CartHandler) HandleCheckoutInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	userID, sessionStoreID, err := h.extractAuthContext(r)
	if err != nil || userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	items, couponCode, err := h.cartStore.GetCart(r.Context(), sessionStoreID, userID)
	if err != nil || len(items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeCartEmpty, "Cannot checkout an empty cart", nil)
		return
	}

	// Lock check & idempotency
	lockAcquired, _ := h.lockManager.AcquireCheckoutLock(r.Context(), userID)
	if !lockAcquired {
		pendingSession, err := h.checkoutRepo.GetPendingSessionByUserID(r.Context(), userID)
		if err == nil && pendingSession != nil {
			// Idempotent retry: return existing active pending session
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(pendingSession)
			return
		}
		errors.WriteError(w, http.StatusConflict, errors.CodeCartLocked, "Active checkout in progress", nil)
		return
	}

	// Re-validate live product prices against Redis/catalog
	for _, item := range items {
		liveProduct, err := h.catalogEngine.GetProductByBarcode(r.Context(), sessionStoreID, item.Barcode)
		if err == nil && liveProduct != nil {
			if math.Abs(float64(liveProduct.PricePaise-item.PricePaiseSnapshot)) > 1.0 { // > ₹0.01 difference
				_ = h.lockManager.ReleaseCheckoutLock(r.Context(), userID)
				errors.WriteError(w, http.StatusConflict, errors.CodePriceChanged, "Price changed for an item in cart. Please review updated cart.", nil)
				return
			}
		}
	}

	// Evaluate offers & GST breakdown
	var subtotalPaise int64
	var itemDTOs []gst.CartItemDTO
	for _, item := range items {
		subtotalPaise += item.PricePaiseSnapshot * int64(item.Qty)
		itemDTOs = append(itemDTOs, gst.CartItemDTO{
			Barcode:        item.Barcode,
			Name:           item.Name,
			Qty:            item.Qty,
			PricePaise:     item.PricePaiseSnapshot,
			LineTotalPaise: item.PricePaiseSnapshot * int64(item.Qty),
			HSNCode:        item.HSNCode,
		})
	}

	discountPaise := int64(0)
	if featureflags.IsEnabled(r.Context(), nil, nil, "offers_engine", sessionStoreID) {
		discountPaise, _, _ = h.offerEngine.EvaluateOffers(r.Context(), sessionStoreID, items, subtotalPaise)
	}

	gstBreakdown := gst.CalculateGST(itemDTOs, "27", "", discountPaise, nil)

	checkoutSession := &CheckoutSession{
		ID:            uuid.New().String(),
		UserID:        userID,
		StoreID:       sessionStoreID,
		Items:         items,
		SubtotalPaise: gstBreakdown.SubtotalPaise,
		DiscountPaise: gstBreakdown.DiscountPaise,
		CGSTPaise:     gstBreakdown.CGSTPaise,
		SGSTPaise:     gstBreakdown.SGSTPaise,
		IGSTPaise:     gstBreakdown.IGSTPaise,
		TotalPaise:    gstBreakdown.TotalPaise,
		CouponCode:    couponCode,
		SupplyType:    gstBreakdown.SupplyType,
		Status:        "PENDING",
	}

	if err := h.checkoutRepo.CreateCheckoutSession(r.Context(), checkoutSession); err != nil {
		_ = h.lockManager.ReleaseCheckoutLock(r.Context(), userID)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create checkout session", nil)
		return
	}

	if h.eventProc != nil {
		h.eventProc.PublishCheckoutInitiated(r.Context(), checkoutSession)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(checkoutSession)
}

func (h *CartHandler) writeCartSummaryResponse(w http.ResponseWriter, r *http.Request, storeID, userID string) {
	items, couponCode, err := h.cartStore.GetCart(r.Context(), storeID, userID)
	if err != nil || items == nil {
		items = []*CartItem{}
	}

	var subtotalPaise int64
	var itemCount int
	var itemDTOs []gst.CartItemDTO

	for _, item := range items {
		itemCount += item.Qty
		lineTotal := item.PricePaiseSnapshot * int64(item.Qty)
		item.LineTotalPaise = lineTotal
		subtotalPaise += lineTotal

		itemDTOs = append(itemDTOs, gst.CartItemDTO{
			Barcode:        item.Barcode,
			Name:           item.Name,
			Qty:            item.Qty,
			PricePaise:     item.PricePaiseSnapshot,
			LineTotalPaise: lineTotal,
			HSNCode:        item.HSNCode,
		})
	}

	discountPaise := int64(0)
	var appliedOffers []string
	if len(items) > 0 && featureflags.IsEnabled(r.Context(), nil, nil, "offers_engine", storeID) {
		discountPaise, appliedOffers, _ = h.offerEngine.EvaluateOffers(r.Context(), storeID, items, subtotalPaise)
	}

	gstBreakdown := gst.CalculateGST(itemDTOs, "27", "", discountPaise, nil)

	summary := &CartSummary{
		Items:         items,
		SubtotalPaise: gstBreakdown.SubtotalPaise,
		DiscountPaise: gstBreakdown.DiscountPaise,
		AppliedOffers: appliedOffers,
		CouponCode:    couponCode,
		CGSTPaise:     gstBreakdown.CGSTPaise,
		SGSTPaise:     gstBreakdown.SGSTPaise,
		IGSTPaise:     gstBreakdown.IGSTPaise,
		TotalPaise:    gstBreakdown.TotalPaise,
		ItemCount:     itemCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *CartHandler) extractAuthContext(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		userID := r.Header.Get("X-User-ID")
		storeID := r.Header.Get("X-Store-ID")
		if userID != "" && storeID != "" {
			return userID, storeID, nil
		}
		return "", "", fmt.Errorf("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return "", "", fmt.Errorf("invalid customer token")
	}

	userID := claims.UserID
	sessionStoreID := claims.StoreID
	if sessionStoreID == "" {
		sessionStoreID = r.Header.Get("X-Store-ID")
	}

	reqStoreID := r.URL.Query().Get("store_id")
	if reqStoreID != "" && sessionStoreID != "" && reqStoreID != sessionStoreID {
		return "", "", fmt.Errorf("cross-store session mismatch")
	}

	return userID, sessionStoreID, nil
}

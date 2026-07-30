package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	repo         Repository
	exitTokenSvc ExitTokenService
	invoiceSvc   InvoiceService
	jwtSecret    string
}

func NewOrderHandler(repo Repository, exitTokenSvc ExitTokenService, invoiceSvc InvoiceService, jwtSecret string) *OrderHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &OrderHandler{
		repo:         repo,
		exitTokenSvc: exitTokenSvc,
		invoiceSvc:   invoiceSvc,
		jwtSecret:    jwtSecret,
	}
}

func (h *OrderHandler) extractAuthClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		userID := r.Header.Get("X-User-ID")
		role := r.Header.Get("X-User-Role")
		storeID := r.Header.Get("X-Store-ID")
		if userID != "" {
			if role == "" {
				role = RoleCustomer
			}
			return &jwt.Claims{
				UserID:  userID,
				Role:    role,
				StoreID: storeID,
			}, nil
		}
		return nil, fmt.Errorf("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// GetOrderHistoryHandler lists customer order history
// @Summary Get Order History
// @Description Paginated order history list for authenticated user
// @Tags Orders
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} sharedErrors.APIError
// @Failure 500 {object} sharedErrors.APIError
// @Router /v1/order/history [get]
func (h *OrderHandler) GetOrderHistoryHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	orders, err := h.repo.GetOrdersByUserID(r.Context(), claims.UserID, page, pageSize)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to query order history", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": orders,
		"page":   page,
	})
}

// GetOrderDetailHandler retrieves full order details
// @Summary Get Order Detail
// @Description Retrieve single order item details, tax breakdowns, and signed invoice URL
// @Tags Orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} sharedErrors.APIError
// @Failure 404 {object} sharedErrors.APIError
// @Router /v1/order/{id} [get]
func (h *OrderHandler) GetOrderDetailHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	if orderID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "order id is required", nil)
		return
	}

	order, err := h.repo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	// Ownership check: user can only view their own order unless SYSTEM/ADMIN
	if order.UserID != claims.UserID && claims.Role != RoleSystem && claims.Role != RoleAdmin {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	// Generate signed short-lived S3 URL for invoice if key is present
	var signedInvoiceURL *string
	if order.InvoiceS3Key != nil {
		url := h.invoiceSvc.GetSignedInvoiceURL(r.Context(), *order.InvoiceS3Key)
		signedInvoiceURL = &url
	}

	resp := map[string]interface{}{
		"order":              order,
		"signed_invoice_url": signedInvoiceURL,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetExitTokenHandler fetches encrypted exit QR token
// @Summary Get Exit Token
// @Description Fetch active exit gate QR token for completed order
// @Tags Orders
// @Produce json
// @Param store_id query string false "Store ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} sharedErrors.APIError
// @Failure 401 {object} sharedErrors.APIError
// @Failure 404 {object} sharedErrors.APIError
// @Router /v1/order/exit-token [get]
func (h *OrderHandler) GetExitTokenHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		storeID = claims.StoreID
	}
	if storeID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "store_id is required", nil)
		return
	}

	exitData, err := h.exitTokenSvc.GetExitToken(r.Context(), claims.UserID, storeID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNoPendingExit, "No active exit pass found or token expired", nil)
		return
	}

	resp := ExitTokenResponse{
		OrderID:   exitData.OrderID,
		ExitToken: exitData.ExitToken,
		ExpiresAt: exitData.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) GetStoreOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		storeID = claims.StoreID
	}
	if storeID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "store_id is required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	orders, err := h.repo.GetOrdersByStoreID(r.Context(), storeID, page, pageSize)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to query store orders", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": orders,
		"page":   page,
	})
}

func (h *OrderHandler) AcceptReturnHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || (claims.Role != RoleAdmin && claims.Role != RoleStoreManager && claims.Role != RoleCashier && claims.Role != "MANAGER" && claims.Role != "STAFF") {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Staff/Manager authorization required", nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	order, err := h.repo.GetOrderByID(r.Context(), orderID)
	if err != nil || order == nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":  order.ID,
		"user_id":   order.UserID,
		"store_id":  order.StoreID,
		"timestamp": time.Now(),
	})

	if err := h.repo.UpdateOrderStatusAndPublishOutbox(r.Context(), order.ID, StatusReturned, TopicOrderReturned, payload); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to accept return request", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": orderID,
		"status":   StatusReturned,
	})
}

func (h *OrderHandler) RejectReturnHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || (claims.Role != RoleAdmin && claims.Role != RoleStoreManager && claims.Role != RoleCashier && claims.Role != "MANAGER" && claims.Role != "STAFF") {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Staff/Manager authorization required", nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	order, err := h.repo.GetOrderByID(r.Context(), orderID)
	if err != nil || order == nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Return request rejected by store manager"
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":  order.ID,
		"user_id":   order.UserID,
		"store_id":  order.StoreID,
		"reason":    req.Reason,
		"timestamp": time.Now(),
	})

	if err := h.repo.UpdateOrderStatusAndPublishOutbox(r.Context(), order.ID, StatusReturnRejected, TopicOrderReturnRejected, payload); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to reject return request", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": orderID,
		"status":   StatusReturnRejected,
		"reason":   req.Reason,
	})
}

func (h *OrderHandler) GetChainOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Authorization required", nil)
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	if claims.UserType == "CHAIN_HQ" {
		if chainID != "" && chainID != claims.ChainID {
			sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Chain isolation violation: cannot query another chain's orders", nil)
			return
		}
		chainID = claims.ChainID
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	orders, total, err := h.repo.ListChainOrders(r.Context(), chainID, page, pageSize)
	if err != nil {
		orders = []*Order{}
		total = 0
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"orders":   orders,
		"total":    total,
		"chain_id": chainID,
	})
}

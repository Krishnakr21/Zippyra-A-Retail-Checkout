package errors

import (
	"encoding/json"
	"net/http"
)

// Canonical Error Codes
const (
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeInternalError       = "INTERNAL_ERROR"
	CodeOTPInvalid          = "OTP_INVALID"
	CodeOTPExpired          = "OTP_EXPIRED"
	CodeOTPLocked           = "OTP_LOCKED"
	CodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
	CodeGoogleTokenInvalid  = "GOOGLE_TOKEN_INVALID"
	CodeGoogleTokenExpired  = "GOOGLE_TOKEN_EXPIRED"
	CodeIdentifierTaken     = "IDENTIFIER_TAKEN"
	CodeStoreNotFound       = "STORE_NOT_FOUND"
	CodeStoreClosed         = "STORE_CLOSED"
	CodeStoreAtCapacity     = "STORE_AT_CAPACITY"
	CodeStoreGeofenceMismatch = "STORE_GEOFENCE_MISMATCH"
	CodeQRTokenInvalid      = "QR_TOKEN_INVALID"
	CodeQRTokenExpired      = "QR_TOKEN_EXPIRED"
	CodeNoActiveSession     = "NO_ACTIVE_SESSION"
	CodeProductNotFound     = "PRODUCT_NOT_FOUND"
	CodeBarcodeInvalid      = "BARCODE_INVALID"
	CodeCategoryNotFound    = "CATEGORY_NOT_FOUND"
	CodeImportFileInvalid   = "IMPORT_FILE_INVALID"
	CodeHSNCodeNotFound     = "HSN_CODE_NOT_FOUND"
	CodeCartLocked          = "CART_LOCKED"
	CodeOutOfStock          = "OUT_OF_STOCK"
	CodeCartEmpty           = "CART_EMPTY"
	CodePriceChanged        = "PRICE_CHANGED"
	CodeCouponInvalid       = "COUPON_INVALID"
	CodeCouponExpired       = "COUPON_EXPIRED"
	CodeCouponMinNotMet     = "COUPON_MIN_NOT_MET"
	CodeCheckoutSessionExpired = "CHECKOUT_SESSION_EXPIRED"
	CodePaymentAlreadyCompleted = "PAYMENT_ALREADY_COMPLETED"
	CodeInsufficientLoyaltyPoints = "INSUFFICIENT_LOYALTY_POINTS"
	CodeInsufficientCash     = "INSUFFICIENT_CASH"
	CodePaymentNotFound      = "PAYMENT_NOT_FOUND"
	CodeGatewayUnavailable   = "GATEWAY_UNAVAILABLE"
	CodeOrderNotFound        = "ORDER_NOT_FOUND"
	CodeNoPendingExit        = "NO_PENDING_EXIT"
	CodeReturnWindowExpired = "RETURN_WINDOW_EXPIRED"
	CodeItemNotReturnable    = "ITEM_NOT_RETURNABLE"
	CodeReturnQtyExceeded    = "RETURN_QTY_EXCEEDED"
	CodeInvalidToken         = "INVALID_TOKEN"
	CodeQRExpired            = "QR_EXPIRED"
	CodeQRAlreadyUsed        = "QR_ALREADY_USED"
	CodeWrongStore           = "WRONG_STORE"
	CodeNotAwaitingRFID      = "NOT_AWAITING_RFID"
	CodeRFIDTimeout          = "RFID_TIMEOUT"
	CodeAccountNotFound      = "ACCOUNT_NOT_FOUND"
	CodeStockNotFound        = "STOCK_NOT_FOUND"
	CodeInsufficientStockForAdjustment = "INSUFFICIENT_STOCK_FOR_ADJUSTMENT"
	CodeInsufficientStockForTransfer   = "INSUFFICIENT_STOCK_FOR_TRANSFER"
	CodePONotReceivable            = "PO_NOT_RECEIVABLE"
	CodePOAlreadySubmitted         = "PO_ALREADY_SUBMITTED"
	CodePOLocked                   = "PO_LOCKED"
	CodeGRNAlreadyCompleted        = "GRN_ALREADY_COMPLETED"
	CodeQCIncomplete               = "QC_INCOMPLETE"
	CodeCrossChainTransferDenied   = "CROSS_CHAIN_TRANSFER_DENIED"
	CodeGSTINChecksumInvalid       = "GSTIN_CHECKSUM_INVALID"
	CodeGSTINStateMismatch         = "GSTIN_STATE_MISMATCH"
	CodeIRNAlreadyIssued           = "IRN_ALREADY_ISSUED"
	CodeDPDPRequestNotFound        = "DPDP_REQUEST_NOT_FOUND"
	CodeKYCRecordNotFound          = "KYC_RECORD_NOT_FOUND"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(code, message string, details interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

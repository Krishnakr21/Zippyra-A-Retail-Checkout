class ErrorCodes {
  static const String otpInvalid = 'OTP_INVALID';
  static const String otpExpired = 'OTP_EXPIRED';
  static const String otpLocked = 'OTP_LOCKED';
  static const String rateLimitExceeded = 'RATE_LIMIT_EXCEEDED';
  static const String googleTokenInvalid = 'GOOGLE_TOKEN_INVALID';
  static const String googleTokenExpired = 'GOOGLE_TOKEN_EXPIRED';
  static const String identifierTaken = 'IDENTIFIER_TAKEN';
  static const String invalidRequest = 'INVALID_REQUEST';
  static const String unauthorized = 'UNAUTHORIZED';

  // Store Session Error Codes
  static const String storeNotFound = 'STORE_NOT_FOUND';
  static const String storeClosed = 'STORE_CLOSED';
  static const String storeAtCapacity = 'STORE_AT_CAPACITY';
  static const String storeGeofenceMismatch = 'STORE_GEOFENCE_MISMATCH';
  static const String qrTokenInvalid = 'QR_TOKEN_INVALID';
  static const String qrTokenExpired = 'QR_TOKEN_EXPIRED';
  static const String noActiveSession = 'NO_ACTIVE_SESSION';

  // Catalog Error Codes
  static const String productNotFound = 'PRODUCT_NOT_FOUND';
  static const String barcodeInvalid = 'BARCODE_INVALID';
  static const String categoryNotFound = 'CATEGORY_NOT_FOUND';
  static const String importFileInvalid = 'IMPORT_FILE_INVALID';
  static const String hsnCodeNotFound = 'HSN_CODE_NOT_FOUND';

  // Cart Error Codes
  static const String cartLocked = 'CART_LOCKED';
  static const String outOfStock = 'OUT_OF_STOCK';
  static const String cartEmpty = 'CART_EMPTY';
  static const String priceChanged = 'PRICE_CHANGED';
  static const String couponInvalid = 'COUPON_INVALID';
  static const String couponExpired = 'COUPON_EXPIRED';
  static const String couponMinNotMet = 'COUPON_MIN_NOT_MET';
  static const String checkoutSessionExpired = 'CHECKOUT_SESSION_EXPIRED';

  // Order Error Codes
  static const String orderNotFound = 'ORDER_NOT_FOUND';
  static const String noPendingExit = 'NO_PENDING_EXIT';
  static const String returnWindowExpired = 'RETURN_WINDOW_EXPIRED';
  static const String itemNotReturnable = 'ITEM_NOT_RETURNABLE';
  static const String returnQtyExceeded = 'RETURN_QTY_EXCEEDED';
}

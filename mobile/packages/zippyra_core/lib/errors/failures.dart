import 'error_codes.dart';

abstract class Failure {
  final String message;
  final String? code;

  const Failure(this.message, {this.code});

  @override
  String toString() => message;
}

class ServerFailure extends Failure {
  const ServerFailure(super.message, {super.code});
}

class AuthFailure extends Failure {
  const AuthFailure(super.message, {super.code});
}

class NetworkFailure extends Failure {
  const NetworkFailure(super.message, {super.code});
}

class CertificatePinningFailure extends NetworkFailure {
  const CertificatePinningFailure([String message = "Can't verify server identity. Check your network connection."])
      : super(message, code: 'CERTIFICATE_PINNING_MISMATCH');
}

class UserCancelledFailure extends Failure {
  const UserCancelledFailure([super.message = 'Sign in cancelled']);
}

// Typed Store Session Failures
class StoreClosedFailure extends Failure {
  const StoreClosedFailure([String message = 'Store is currently closed'])
      : super(message, code: 'STORE_CLOSED');
}

class StoreAtCapacityFailure extends Failure {
  const StoreAtCapacityFailure([String message = 'Store is at maximum capacity'])
      : super(message, code: 'STORE_AT_CAPACITY');
}

class StoreGeofenceMismatchFailure extends Failure {
  const StoreGeofenceMismatchFailure([String message = 'You must be at the store entrance to enter'])
      : super(message, code: 'STORE_GEOFENCE_MISMATCH');
}

class QRTokenInvalidFailure extends Failure {
  const QRTokenInvalidFailure([String message = 'Invalid entrance QR token'])
      : super(message, code: 'QR_TOKEN_INVALID');
}

class QRTokenExpiredFailure extends Failure {
  const QRTokenExpiredFailure([String message = 'Entrance QR token has expired'])
      : super(message, code: 'QR_TOKEN_EXPIRED');
}

class NoActiveSessionFailure extends Failure {
  const NoActiveSessionFailure([String message = 'No active store session found'])
      : super(message, code: 'NO_ACTIVE_SESSION');
}

// Typed Cart Failures
class BarcodeInvalidFailure extends Failure {
  const BarcodeInvalidFailure([String message = 'Invalid barcode'])
      : super(message, code: ErrorCodes.barcodeInvalid);
}

class ProductNotFoundFailure extends Failure {
  const ProductNotFoundFailure([String message = 'Product not found'])
      : super(message, code: ErrorCodes.productNotFound);
}

class OutOfStockFailure extends Failure {
  const OutOfStockFailure([String message = 'Item is out of stock'])
      : super(message, code: ErrorCodes.outOfStock);
}

class CartLockedFailure extends Failure {
  const CartLockedFailure([String message = 'Checkout in progress'])
      : super(message, code: ErrorCodes.cartLocked);
}

class CartEmptyFailure extends Failure {
  const CartEmptyFailure([String message = 'Cart is empty'])
      : super(message, code: ErrorCodes.cartEmpty);
}

class CouponInvalidFailure extends Failure {
  const CouponInvalidFailure([String message = 'Invalid coupon code'])
      : super(message, code: ErrorCodes.couponInvalid);
}

class CouponExpiredFailure extends Failure {
  const CouponExpiredFailure([String message = 'Coupon has expired'])
      : super(message, code: ErrorCodes.couponExpired);
}

class CouponMinNotMetFailure extends Failure {
  const CouponMinNotMetFailure([String message = 'Minimum cart value not met'])
      : super(message, code: ErrorCodes.couponMinNotMet);
}

class PriceChangedFailure extends Failure {
  final dynamic updatedCart;
  const PriceChangedFailure([this.updatedCart, String message = 'Item prices have changed. Please review your updated cart.'])
      : super(message, code: ErrorCodes.priceChanged);
}

// Typed Order Failures
class OrderNotFoundFailure extends Failure {
  const OrderNotFoundFailure([String message = 'Order not found'])
      : super(message, code: ErrorCodes.orderNotFound);
}

class NoPendingExitFailure extends Failure {
  const NoPendingExitFailure([String message = 'No pending exit pass found'])
      : super(message, code: ErrorCodes.noPendingExit);
}

class ReturnWindowExpiredFailure extends Failure {
  const ReturnWindowExpiredFailure([String message = 'Return window of 24 hours has expired for this order'])
      : super(message, code: ErrorCodes.returnWindowExpired);
}

class ItemNotReturnableFailure extends Failure {
  const ItemNotReturnableFailure([String message = 'Item is not returnable'])
      : super(message, code: ErrorCodes.itemNotReturnable);
}

class ReturnQtyExceededFailure extends Failure {
  const ReturnQtyExceededFailure([String message = 'Return quantity exceeds purchased quantity'])
      : super(message, code: ErrorCodes.returnQtyExceeded);
}




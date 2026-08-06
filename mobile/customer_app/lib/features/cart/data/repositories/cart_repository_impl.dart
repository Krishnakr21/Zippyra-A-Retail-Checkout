import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/cart_summary.dart';
import '../../domain/entities/checkout_session.dart';
import '../../domain/repositories/cart_repository.dart';
import '../datasources/cart_remote_data_source.dart';
import '../models/cart_summary_model.dart';

class CartRepositoryImpl implements CartRepository {
  final CartRemoteDataSource remoteDataSource;
  CartSummary? _cachedSummary;

  CartRepositoryImpl({required this.remoteDataSource});

  @override
  CartSummary? getCachedCart() => _cachedSummary;

  @override
  void clearMemoryCache() {
    _cachedSummary = null;
  }

  @override
  Future<CartSummary> getCart(String storeId) async {
    try {
      final summary = await remoteDataSource.getCart(storeId);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> scanItem(String storeId, String barcode, {int qty = 1}) async {
    try {
      final summary = await remoteDataSource.scanItem(storeId, barcode, qty: qty);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> updateQuantity(String storeId, String barcode, int qty) async {
    try {
      final summary = await remoteDataSource.updateQuantity(storeId, barcode, qty);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> removeItem(String storeId, String barcode) async {
    try {
      final summary = await remoteDataSource.removeItem(storeId, barcode);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> clearCart(String storeId) async {
    try {
      final summary = await remoteDataSource.clearCart(storeId);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> applyCoupon(String storeId, String code) async {
    try {
      final summary = await remoteDataSource.applyCoupon(storeId, code);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CartSummary> removeCoupon(String storeId) async {
    try {
      final summary = await remoteDataSource.removeCoupon(storeId);
      _cachedSummary = summary;
      return summary;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  @override
  Future<CheckoutSession> initCheckout(String storeId) async {
    try {
      final session = await remoteDataSource.initCheckout(storeId);
      return session;
    } catch (e) {
      throw _mapExceptionToFailure(e);
    }
  }

  Failure _mapExceptionToFailure(Object e) {
    if (e is Failure) return e;

    if (e is DioException && e.response != null) {
      final status = e.response!.statusCode;
      final data = e.response!.data;
      String? code;
      String message = 'An error occurred';

      if (data is Map<String, dynamic> && data.containsKey('error')) {
        final errObj = data['error'] as Map<String, dynamic>;
        code = errObj['code'] as String?;
        message = errObj['message'] as String? ?? message;
      }

      switch (code) {
        case ErrorCodes.barcodeInvalid:
          return BarcodeInvalidFailure(message);
        case ErrorCodes.productNotFound:
          return ProductNotFoundFailure(message);
        case ErrorCodes.outOfStock:
          return OutOfStockFailure(message);
        case ErrorCodes.cartLocked:
          return CartLockedFailure(message);
        case ErrorCodes.cartEmpty:
          return CartEmptyFailure(message);
        case ErrorCodes.priceChanged:
          CartSummary? updatedCart;
          if (data is Map<String, dynamic> && data.containsKey('cart')) {
            try {
              updatedCart = CartSummaryModel.fromJson(data['cart'] as Map<String, dynamic>);
            } catch (_) {}
          }
          return PriceChangedFailure(updatedCart, message);
        case ErrorCodes.couponInvalid:
          return CouponInvalidFailure(message);
        case ErrorCodes.couponExpired:
          return CouponExpiredFailure(message);
        case ErrorCodes.couponMinNotMet:
          return CouponMinNotMetFailure(message);
      }

      if (status == 409) {
        if (code == ErrorCodes.cartLocked) return CartLockedFailure(message);
        return OutOfStockFailure(message);
      }
      return ServerFailure(message, code: code);
    }

    return ServerFailure(e.toString());
  }
}

import '../entities/cart_summary.dart';
import '../entities/checkout_session.dart';

abstract class CartRepository {
  Future<CartSummary> getCart(String storeId);
  Future<CartSummary> scanItem(String storeId, String barcode, {int qty = 1});
  Future<CartSummary> updateQuantity(String storeId, String barcode, int qty);
  Future<CartSummary> removeItem(String storeId, String barcode);
  Future<CartSummary> clearCart(String storeId);
  Future<CartSummary> applyCoupon(String storeId, String code);
  Future<CartSummary> removeCoupon(String storeId);
  Future<CheckoutSession> initCheckout(String storeId);
  CartSummary? getCachedCart();
  void clearMemoryCache();
}

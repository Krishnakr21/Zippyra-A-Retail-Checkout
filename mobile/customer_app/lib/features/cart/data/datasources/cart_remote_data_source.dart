import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../models/cart_summary_model.dart';
import '../models/checkout_session_model.dart';

abstract class CartRemoteDataSource {
  Future<CartSummaryModel> getCart(String storeId);
  Future<CartSummaryModel> scanItem(String storeId, String barcode, {int qty = 1});
  Future<CartSummaryModel> updateQuantity(String storeId, String barcode, int qty);
  Future<CartSummaryModel> removeItem(String storeId, String barcode);
  Future<CartSummaryModel> clearCart(String storeId);
  Future<CartSummaryModel> applyCoupon(String storeId, String code);
  Future<CartSummaryModel> removeCoupon(String storeId);
  Future<CheckoutSessionModel> initCheckout(String storeId);
}

class CartRemoteDataSourceImpl implements CartRemoteDataSource {
  final ApiClient apiClient;

  CartRemoteDataSourceImpl({required this.apiClient});

  Options _getOptions(String storeId) {
    return Options(
      headers: {
        'X-User-ID': 'user-demo-1',
        'X-Store-ID': storeId.isNotEmpty ? storeId : 'store-1',
      },
    );
  }

  String _getUrl(String path) {
    if (path.startsWith('http')) return path;
    return 'http://localhost:8084$path';
  }

  @override
  Future<CartSummaryModel> getCart(String storeId) async {
    final url = _getUrl('/v1/cart?store_id=$storeId');
    final response = await apiClient.dio.get(url, options: _getOptions(storeId));
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> scanItem(String storeId, String barcode, {int qty = 1}) async {
    final url = _getUrl('/v1/cart/scan?store_id=$storeId');
    final response = await apiClient.dio.post(
      url,
      data: {'barcode': barcode, 'qty': qty},
      options: _getOptions(storeId),
    );
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> updateQuantity(String storeId, String barcode, int qty) async {
    final url = _getUrl('/v1/cart/item/$barcode?store_id=$storeId');
    final response = await apiClient.dio.put(
      url,
      data: {'qty': qty},
      options: _getOptions(storeId),
    );
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> removeItem(String storeId, String barcode) async {
    final url = _getUrl('/v1/cart/item/$barcode?store_id=$storeId');
    final response = await apiClient.dio.delete(url, options: _getOptions(storeId));
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> clearCart(String storeId) async {
    final url = _getUrl('/v1/cart?store_id=$storeId');
    final response = await apiClient.dio.delete(url, options: _getOptions(storeId));
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> applyCoupon(String storeId, String code) async {
    final url = _getUrl('/v1/cart/coupon/apply?store_id=$storeId');
    final response = await apiClient.dio.post(
      url,
      data: {'code': code},
      options: _getOptions(storeId),
    );
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CartSummaryModel> removeCoupon(String storeId) async {
    final url = _getUrl('/v1/cart/coupon?store_id=$storeId');
    final response = await apiClient.dio.delete(url, options: _getOptions(storeId));
    return CartSummaryModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<CheckoutSessionModel> initCheckout(String storeId) async {
    final url = _getUrl('/v1/cart/checkout/init?store_id=$storeId');
    final response = await apiClient.dio.post(url, options: _getOptions(storeId));
    return CheckoutSessionModel.fromJson(response.data as Map<String, dynamic>);
  }
}

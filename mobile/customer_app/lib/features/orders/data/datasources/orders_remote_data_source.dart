import 'package:zippyra_core/zippyra_core.dart';
import '../models/exit_token_model.dart';
import '../models/order_detail_model.dart';
import '../models/order_summary_model.dart';

abstract class OrdersRemoteDataSource {
  Future<List<OrderSummaryModel>> getOrderHistory({int page = 1, int pageSize = 20});
  Future<OrderDetailModel> getOrderDetail(String orderId);
  Future<void> requestReturn({
    required String orderId,
    required List<String> itemBarcodes,
    required String reason,
  });
  Future<ExitTokenModel> getExitToken({required String storeId});
}

class OrdersRemoteDataSourceImpl implements OrdersRemoteDataSource {
  final ApiClient apiClient;

  OrdersRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<OrderSummaryModel>> getOrderHistory({int page = 1, int pageSize = 20}) async {
    try {
      final response = await apiClient.get('/v1/order/history', queryParameters: {
        'page': page,
        'page_size': pageSize,
      });
      final data = response.data as Map<String, dynamic>;
      final list = (data['orders'] as List<dynamic>?)
              ?.map((e) => OrderSummaryModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [];
      return list;
    } catch (_) {
      return [];
    }
  }

  @override
  Future<OrderDetailModel> getOrderDetail(String orderId) async {
    final response = await apiClient.get('/v1/order/$orderId');
    final data = response.data as Map<String, dynamic>;
    return OrderDetailModel.fromJson(data);
  }

  @override
  Future<void> requestReturn({
    required String orderId,
    required List<String> itemBarcodes,
    required String reason,
  }) async {
    await apiClient.post(
      '/v1/order/$orderId/return',
      data: {
        'item_barcodes': itemBarcodes,
        'reason': reason,
      },
    );
  }

  @override
  Future<ExitTokenModel> getExitToken({required String storeId}) async {
    final response = await apiClient.get('/v1/order/exit-token', queryParameters: {
      'store_id': storeId,
    });
    final data = response.data as Map<String, dynamic>;
    return ExitTokenModel.fromJson(data);
  }
}

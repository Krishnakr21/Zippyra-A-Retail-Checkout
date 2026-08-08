import '../entities/exit_token.dart';
import '../entities/order_detail.dart';
import '../entities/order_summary.dart';

abstract class OrdersRepository {
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20});
  Future<OrderDetail> getOrderDetail(String orderId);
  Future<void> requestReturn({
    required String orderId,
    required List<String> itemBarcodes,
    required String reason,
  });
  Future<ExitToken> getExitToken({required String storeId});
}

import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/exit_token.dart';
import '../../domain/entities/order_detail.dart';
import '../../domain/entities/order_summary.dart';
import '../../domain/repositories/orders_repository.dart';
import '../datasources/orders_remote_data_source.dart';

class OrdersRepositoryImpl implements OrdersRepository {
  final OrdersRemoteDataSource remoteDataSource;

  OrdersRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20}) async {
    try {
      return await remoteDataSource.getOrderHistory(page: page, pageSize: pageSize);
    } catch (e) {
      throw ServerFailure('Failed to fetch order history: $e');
    }
  }

  @override
  Future<OrderDetail> getOrderDetail(String orderId) async {
    try {
      return await remoteDataSource.getOrderDetail(orderId);
    } catch (e) {
      final msg = e.toString();
      if (msg.contains(ErrorCodes.orderNotFound)) {
        throw const OrderNotFoundFailure();
      }
      throw ServerFailure('Failed to fetch order detail: $e');
    }
  }

  @override
  Future<void> requestReturn({
    required String orderId,
    required List<String> itemBarcodes,
    required String reason,
  }) async {
    try {
      await remoteDataSource.requestReturn(
        orderId: orderId,
        itemBarcodes: itemBarcodes,
        reason: reason,
      );
    } catch (e) {
      final msg = e.toString();
      if (msg.contains(ErrorCodes.returnWindowExpired)) {
        throw const ReturnWindowExpiredFailure();
      } else if (msg.contains(ErrorCodes.itemNotReturnable)) {
        throw const ItemNotReturnableFailure();
      } else if (msg.contains(ErrorCodes.returnQtyExceeded)) {
        throw const ReturnQtyExceededFailure();
      }
      throw ServerFailure('Failed to submit return request: $e');
    }
  }

  @override
  Future<ExitToken> getExitToken({required String storeId}) async {
    try {
      return await remoteDataSource.getExitToken(storeId: storeId);
    } catch (e) {
      final msg = e.toString();
      if (msg.contains(ErrorCodes.noPendingExit)) {
        throw const NoPendingExitFailure();
      }
      throw ServerFailure('Failed to fetch exit token: $e');
    }
  }
}

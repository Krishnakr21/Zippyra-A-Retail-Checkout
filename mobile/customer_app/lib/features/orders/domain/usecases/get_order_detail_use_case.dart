import '../entities/order_detail.dart';
import '../repositories/orders_repository.dart';

class GetOrderDetailUseCase {
  final OrdersRepository repository;

  GetOrderDetailUseCase(this.repository);

  Future<OrderDetail> call(String orderId) {
    return repository.getOrderDetail(orderId);
  }
}

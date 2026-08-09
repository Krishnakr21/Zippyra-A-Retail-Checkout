import '../entities/order_summary.dart';
import '../repositories/orders_repository.dart';

class GetOrderHistoryUseCase {
  final OrdersRepository repository;

  GetOrderHistoryUseCase(this.repository);

  Future<List<OrderSummary>> call({int page = 1, int pageSize = 20}) {
    return repository.getOrderHistory(page: page, pageSize: pageSize);
  }
}

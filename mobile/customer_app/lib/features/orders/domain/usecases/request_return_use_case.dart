import '../repositories/orders_repository.dart';

class RequestReturnUseCase {
  final OrdersRepository repository;

  RequestReturnUseCase(this.repository);

  Future<void> call({
    required String orderId,
    required List<String> itemBarcodes,
    required String reason,
  }) {
    return repository.requestReturn(
      orderId: orderId,
      itemBarcodes: itemBarcodes,
      reason: reason,
    );
  }
}

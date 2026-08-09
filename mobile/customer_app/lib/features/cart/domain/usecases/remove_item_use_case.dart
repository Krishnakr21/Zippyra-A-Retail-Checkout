import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class RemoveItemUseCase {
  final CartRepository repository;

  RemoveItemUseCase(this.repository);

  Future<CartSummary> call(String storeId, String barcode) async {
    return await repository.removeItem(storeId, barcode);
  }
}

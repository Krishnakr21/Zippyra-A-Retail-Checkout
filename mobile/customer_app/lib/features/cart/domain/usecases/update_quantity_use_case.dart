import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class UpdateQuantityUseCase {
  final CartRepository repository;

  UpdateQuantityUseCase(this.repository);

  Future<CartSummary> call(String storeId, String barcode, int qty) async {
    return await repository.updateQuantity(storeId, barcode, qty);
  }
}

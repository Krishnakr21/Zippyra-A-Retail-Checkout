import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class ClearCartUseCase {
  final CartRepository repository;

  ClearCartUseCase(this.repository);

  Future<CartSummary> call(String storeId) async {
    return await repository.clearCart(storeId);
  }
}

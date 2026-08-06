import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class GetCartUseCase {
  final CartRepository repository;

  GetCartUseCase(this.repository);

  Future<CartSummary> call(String storeId) async {
    return await repository.getCart(storeId);
  }
}

import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class RemoveCouponUseCase {
  final CartRepository repository;

  RemoveCouponUseCase(this.repository);

  Future<CartSummary> call(String storeId) async {
    return await repository.removeCoupon(storeId);
  }
}

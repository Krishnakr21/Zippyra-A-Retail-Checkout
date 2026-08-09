import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class ApplyCouponUseCase {
  final CartRepository repository;

  ApplyCouponUseCase(this.repository);

  Future<CartSummary> call(String storeId, String code) async {
    return await repository.applyCoupon(storeId, code);
  }
}

import '../entities/checkout_session.dart';
import '../repositories/cart_repository.dart';

class InitCheckoutUseCase {
  final CartRepository repository;

  InitCheckoutUseCase(this.repository);

  Future<CheckoutSession> call(String storeId) async {
    return await repository.initCheckout(storeId);
  }
}

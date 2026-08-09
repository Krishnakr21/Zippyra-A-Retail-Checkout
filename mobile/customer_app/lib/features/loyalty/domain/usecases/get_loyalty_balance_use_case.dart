import '../entities/loyalty_balance.dart';
import '../repositories/loyalty_repository.dart';

class GetLoyaltyBalanceUseCase {
  final LoyaltyRepository repository;

  GetLoyaltyBalanceUseCase(this.repository);

  Future<LoyaltyBalance> call() {
    return repository.getLoyaltyBalance();
  }
}

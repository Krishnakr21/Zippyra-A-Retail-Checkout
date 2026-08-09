import '../entities/loyalty_ledger_entry.dart';
import '../repositories/loyalty_repository.dart';

class GetLoyaltyHistoryUseCase {
  final LoyaltyRepository repository;

  GetLoyaltyHistoryUseCase(this.repository);

  Future<List<LoyaltyLedgerEntry>> call({int page = 1, int pageSize = 20}) {
    return repository.getLoyaltyHistory(page: page, pageSize: pageSize);
  }
}

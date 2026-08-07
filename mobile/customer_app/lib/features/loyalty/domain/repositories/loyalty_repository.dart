import '../entities/loyalty_balance.dart';
import '../entities/loyalty_ledger_entry.dart';
import '../entities/tier_info.dart';

abstract class LoyaltyRepository {
  Future<LoyaltyBalance> getLoyaltyBalance();
  Future<List<LoyaltyLedgerEntry>> getLoyaltyHistory({int page = 1, int pageSize = 20});
  Future<List<TierInfo>> getTiersInfo();
}

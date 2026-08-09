import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_balance.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_ledger_entry.dart';
import 'package:customer_app/features/loyalty/domain/entities/tier_info.dart';
import 'package:customer_app/features/loyalty/domain/repositories/loyalty_repository.dart';
import 'package:customer_app/features/loyalty/domain/usecases/get_tiers_info_use_case.dart';

class MockTiersRepo implements LoyaltyRepository {
  int fetchCallCount = 0;

  @override
  Future<List<TierInfo>> getTiersInfo() async {
    fetchCallCount++;
    return const [
      TierInfo(tier: 'BRONZE', displayName: 'Bronze Tier', minLifetimePoints: 0, earnMultiplier: 1.0),
      TierInfo(tier: 'SILVER', displayName: 'Silver Tier', minLifetimePoints: 5000, earnMultiplier: 1.2),
      TierInfo(tier: 'GOLD', displayName: 'Gold Tier', minLifetimePoints: 20000, earnMultiplier: 1.5),
      TierInfo(tier: 'PLATINUM', displayName: 'Platinum Tier', minLifetimePoints: 50000, earnMultiplier: 2.0),
    ];
  }

  @override
  Future<LoyaltyBalance> getLoyaltyBalance() async => throw UnimplementedError();

  @override
  Future<List<LoyaltyLedgerEntry>> getLoyaltyHistory({int page = 1, int pageSize = 20}) async => throw UnimplementedError();
}

void main() {
  test('GetTiersInfoUseCase serves from local cache on 2nd call within TTL window', () async {
    final repo = MockTiersRepo();
    final useCase = GetTiersInfoUseCase(repo);

    // Call 1 -> Remote fetch
    final result1 = await useCase();
    expect(result1.isNotEmpty, true);
    expect(repo.fetchCallCount, equals(1));

    // Call 2 -> Served from cache
    final result2 = await useCase();
    expect(result2.isNotEmpty, true);
    expect(repo.fetchCallCount, equals(1)); // No 2nd remote call made!
  });
}

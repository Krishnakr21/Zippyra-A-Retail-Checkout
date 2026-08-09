import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_balance.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_ledger_entry.dart';
import 'package:customer_app/features/loyalty/domain/entities/tier_info.dart';
import 'package:customer_app/features/loyalty/domain/repositories/loyalty_repository.dart';
import 'package:customer_app/features/loyalty/domain/usecases/get_loyalty_balance_use_case.dart';
import 'package:customer_app/features/loyalty/presentation/bloc/loyalty_bloc.dart';

class MockLoyaltyRepo implements LoyaltyRepository {
  LoyaltyBalance? balanceToReturn;
  Exception? exceptionToReturn;

  @override
  Future<LoyaltyBalance> getLoyaltyBalance() async {
    if (exceptionToReturn != null) {
      throw exceptionToReturn!;
    }
    return balanceToReturn ??
        const LoyaltyBalance(
          pointsBalance: 1250,
          pointsReserved: 0,
          tier: 'GOLD',
          tierDisplayName: 'Gold Tier',
          lifetimePointsEarned: 22500,
          pointsToNextTier: 27500,
          nextTierName: 'Platinum Tier',
        );
  }

  @override
  Future<List<LoyaltyLedgerEntry>> getLoyaltyHistory({int page = 1, int pageSize = 20}) async {
    return const [];
  }

  @override
  Future<List<TierInfo>> getTiersInfo() async {
    return const [];
  }
}

void main() {
  late MockLoyaltyRepo repo;
  late LoyaltyBloc bloc;

  setUp(() {
    repo = MockLoyaltyRepo();
    bloc = LoyaltyBloc(getLoyaltyBalanceUseCase: GetLoyaltyBalanceUseCase(repo));
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is LoyaltyInitial', () {
    expect(bloc.state, isA<LoyaltyInitial>());
  });

  blocTest<LoyaltyBloc, LoyaltyState>(
    'LoyaltyBalanceRequested success populates all fields correctly',
    build: () => bloc,
    act: (b) => b.add(const LoyaltyBalanceRequested()),
    expect: () => [
      isA<LoyaltyLoading>(),
      isA<LoyaltyLoaded>()
          .having((s) => s.pointsBalance, 'pointsBalance', 1250)
          .having((s) => s.tier, 'tier', 'GOLD')
          .having((s) => s.nextTierName, 'nextTierName', 'Platinum Tier'),
    ],
  );

  blocTest<LoyaltyBloc, LoyaltyState>(
    'Top tier (PLATINUM) returns null pointsToNextTier and null nextTierName',
    build: () {
      repo.balanceToReturn = const LoyaltyBalance(
        pointsBalance: 55000,
        pointsReserved: 0,
        tier: 'PLATINUM',
        tierDisplayName: 'Platinum Tier',
        lifetimePointsEarned: 55000,
        pointsToNextTier: null,
        nextTierName: null,
      );
      return bloc;
    },
    act: (b) => b.add(const LoyaltyBalanceRequested()),
    expect: () => [
      isA<LoyaltyLoading>(),
      isA<LoyaltyLoaded>()
          .having((s) => s.tier, 'tier', 'PLATINUM')
          .having((s) => s.pointsToNextTier, 'pointsToNextTier', isNull)
          .having((s) => s.nextTierName, 'nextTierName', isNull),
    ],
  );
}

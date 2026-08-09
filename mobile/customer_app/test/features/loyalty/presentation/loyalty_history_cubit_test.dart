import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_balance.dart';
import 'package:customer_app/features/loyalty/domain/entities/loyalty_ledger_entry.dart';
import 'package:customer_app/features/loyalty/domain/entities/tier_info.dart';
import 'package:customer_app/features/loyalty/domain/repositories/loyalty_repository.dart';
import 'package:customer_app/features/loyalty/domain/usecases/get_loyalty_history_use_case.dart';
import 'package:customer_app/features/loyalty/presentation/cubit/loyalty_history_cubit.dart';

class MockLoyaltyHistoryRepo implements LoyaltyRepository {
  @override
  Future<List<LoyaltyLedgerEntry>> getLoyaltyHistory({int page = 1, int pageSize = 20}) async {
    if (page == 1) {
      return List.generate(
        20,
        (i) => LoyaltyLedgerEntry(
          entryType: 'EARN',
          pointsDelta: 10,
          createdAt: DateTime.now(),
          balanceAfter: (i + 1) * 10,
        ),
      );
    } else if (page == 2) {
      return List.generate(
        5,
        (i) => LoyaltyLedgerEntry(
          entryType: 'EARN',
          pointsDelta: 10,
          createdAt: DateTime.now(),
          balanceAfter: (20 + i + 1) * 10,
        ),
      );
    }
    return const [];
  }

  @override
  Future<LoyaltyBalance> getLoyaltyBalance() async {
    throw UnimplementedError();
  }

  @override
  Future<List<TierInfo>> getTiersInfo() async {
    throw UnimplementedError();
  }
}

void main() {
  late MockLoyaltyHistoryRepo repo;
  late LoyaltyHistoryCubit cubit;

  setUp(() {
    repo = MockLoyaltyHistoryRepo();
    cubit = LoyaltyHistoryCubit(getLoyaltyHistoryUseCase: GetLoyaltyHistoryUseCase(repo));
  });

  tearDown(() {
    cubit.close();
  });

  test('initial state is LoyaltyHistoryInitial', () {
    expect(cubit.state, isA<LoyaltyHistoryInitial>());
  });

  blocTest<LoyaltyHistoryCubit, LoyaltyHistoryState>(
    'fetchHistory and fetchNextPage pagination appends entries without duplicates',
    build: () => cubit,
    act: (c) async {
      await c.fetchHistory();
      await c.fetchNextPage();
    },
    expect: () => [
      isA<LoyaltyHistoryLoading>(),
      isA<LoyaltyHistoryLoaded>()
          .having((s) => s.items.length, 'page 1 count', 20)
          .having((s) => s.hasMore, 'hasMore page 1', true),
      isA<LoyaltyHistoryLoaded>()
          .having((s) => s.items.length, 'page 2 count', 25)
          .having((s) => s.hasMore, 'hasMore page 2', false),
    ],
  );
}

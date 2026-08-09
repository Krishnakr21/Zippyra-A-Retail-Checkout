import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/loyalty_balance.dart';
import '../../domain/entities/loyalty_ledger_entry.dart';
import '../../domain/entities/tier_info.dart';
import '../../domain/repositories/loyalty_repository.dart';
import '../datasources/loyalty_remote_data_source.dart';

class LoyaltyRepositoryImpl implements LoyaltyRepository {
  final LoyaltyRemoteDataSource remoteDataSource;

  LoyaltyRepositoryImpl({required this.remoteDataSource});

  @override
  Future<LoyaltyBalance> getLoyaltyBalance() async {
    try {
      return await remoteDataSource.getLoyaltyBalance();
    } catch (e) {
      if (e.toString().contains('ACCOUNT_NOT_FOUND')) {
        return const LoyaltyBalance(
          pointsBalance: 0,
          pointsReserved: 0,
          tier: 'BRONZE',
          tierDisplayName: 'Bronze Tier',
          lifetimePointsEarned: 0,
          pointsToNextTier: 5000,
          nextTierName: 'Silver Tier',
        );
      }
      throw ServerFailure('Failed to fetch loyalty balance: $e');
    }
  }

  @override
  Future<List<LoyaltyLedgerEntry>> getLoyaltyHistory({int page = 1, int pageSize = 20}) async {
    try {
      return await remoteDataSource.getLoyaltyHistory(page: page, pageSize: pageSize);
    } catch (e) {
      if (e.toString().contains('ACCOUNT_NOT_FOUND')) {
        return const [];
      }
      throw ServerFailure('Failed to fetch loyalty history: $e');
    }
  }

  @override
  Future<List<TierInfo>> getTiersInfo() async {
    try {
      return await remoteDataSource.getTiersInfo();
    } catch (e) {
      throw ServerFailure('Failed to fetch tier info: $e');
    }
  }
}

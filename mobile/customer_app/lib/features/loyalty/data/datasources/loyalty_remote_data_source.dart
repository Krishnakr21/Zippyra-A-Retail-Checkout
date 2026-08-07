import 'package:zippyra_core/zippyra_core.dart';
import '../models/loyalty_balance_model.dart';
import '../models/loyalty_ledger_entry_model.dart';
import '../models/tier_info_model.dart';

abstract class LoyaltyRemoteDataSource {
  Future<LoyaltyBalanceModel> getLoyaltyBalance();
  Future<List<LoyaltyLedgerEntryModel>> getLoyaltyHistory({int page = 1, int pageSize = 20});
  Future<List<TierInfoModel>> getTiersInfo();
}

class LoyaltyRemoteDataSourceImpl implements LoyaltyRemoteDataSource {
  final ApiClient apiClient;

  LoyaltyRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<LoyaltyBalanceModel> getLoyaltyBalance() async {
    final response = await apiClient.get('/v1/loyalty/balance');
    final data = response.data as Map<String, dynamic>;
    return LoyaltyBalanceModel.fromJson(data);
  }

  @override
  Future<List<LoyaltyLedgerEntryModel>> getLoyaltyHistory({int page = 1, int pageSize = 20}) async {
    final response = await apiClient.get(
      '/v1/loyalty/history',
      queryParameters: {'page': page, 'page_size': pageSize},
    );
    final data = response.data as Map<String, dynamic>;
    final items = data['items'] as List<dynamic>? ?? [];
    return items.map((item) => LoyaltyLedgerEntryModel.fromJson(item as Map<String, dynamic>)).toList();
  }

  @override
  Future<List<TierInfoModel>> getTiersInfo() async {
    final response = await apiClient.get('/v1/loyalty/tiers');
    final data = response.data as Map<String, dynamic>;
    final tiers = data['tiers'] as List<dynamic>? ?? [];
    return tiers.map((t) => TierInfoModel.fromJson(t as Map<String, dynamic>)).toList();
  }
}

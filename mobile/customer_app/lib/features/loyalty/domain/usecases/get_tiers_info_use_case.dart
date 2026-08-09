import '../entities/tier_info.dart';
import '../repositories/loyalty_repository.dart';

class GetTiersInfoUseCase {
  final LoyaltyRepository repository;

  List<TierInfo>? _cachedTiers;
  DateTime? _lastFetchTime;
  static const Duration _cacheTtl = Duration(hours: 24);

  GetTiersInfoUseCase(this.repository);

  Future<List<TierInfo>> call({bool forceRefresh = false}) async {
    final now = DateTime.now();
    if (!forceRefresh && _cachedTiers != null && _lastFetchTime != null) {
      if (now.difference(_lastFetchTime!) < _cacheTtl) {
        return _cachedTiers!;
      }
    }

    try {
      final tiers = await repository.getTiersInfo();
      _cachedTiers = tiers;
      _lastFetchTime = now;
      return tiers;
    } catch (e) {
      if (_cachedTiers != null) return _cachedTiers!;
      rethrow;
    }
  }
}

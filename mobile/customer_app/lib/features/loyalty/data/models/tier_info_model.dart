import '../../domain/entities/tier_info.dart';

class TierInfoModel extends TierInfo {
  const TierInfoModel({
    required super.tier,
    required super.displayName,
    required super.minLifetimePoints,
    required super.earnMultiplier,
  });

  factory TierInfoModel.fromJson(Map<String, dynamic> json) {
    return TierInfoModel(
      tier: json['tier'] as String? ?? 'BRONZE',
      displayName: json['display_name'] as String? ?? 'Bronze Tier',
      minLifetimePoints: (json['min_lifetime_points'] as num?)?.toInt() ?? 0,
      earnMultiplier: (json['earn_multiplier'] as num?)?.toDouble() ?? 1.0,
    );
  }
}

import '../../domain/entities/loyalty_balance.dart';

class LoyaltyBalanceModel extends LoyaltyBalance {
  const LoyaltyBalanceModel({
    required super.pointsBalance,
    required super.pointsReserved,
    required super.tier,
    required super.tierDisplayName,
    required super.lifetimePointsEarned,
    super.pointsToNextTier,
    super.nextTierName,
  });

  factory LoyaltyBalanceModel.fromJson(Map<String, dynamic> json) {
    return LoyaltyBalanceModel(
      pointsBalance: (json['points_balance'] as num?)?.toInt() ?? 0,
      pointsReserved: (json['points_reserved'] as num?)?.toInt() ?? 0,
      tier: json['tier'] as String? ?? 'BRONZE',
      tierDisplayName: json['tier_display_name'] as String? ?? 'Bronze Tier',
      lifetimePointsEarned: (json['lifetime_points_earned'] as num?)?.toInt() ?? 0,
      pointsToNextTier: (json['points_to_next_tier'] as num?)?.toInt(),
      nextTierName: json['next_tier_name'] as String?,
    );
  }
}

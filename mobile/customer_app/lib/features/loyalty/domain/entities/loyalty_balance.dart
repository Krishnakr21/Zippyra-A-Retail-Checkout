import 'package:equatable/equatable.dart';

class LoyaltyBalance extends Equatable {
  final int pointsBalance;
  final int pointsReserved;
  final String tier;
  final String tierDisplayName;
  final int lifetimePointsEarned;
  final int? pointsToNextTier;
  final String? nextTierName;

  const LoyaltyBalance({
    required this.pointsBalance,
    required this.pointsReserved,
    required this.tier,
    required this.tierDisplayName,
    required this.lifetimePointsEarned,
    this.pointsToNextTier,
    this.nextTierName,
  });

  @override
  List<Object?> get props => [
        pointsBalance,
        pointsReserved,
        tier,
        tierDisplayName,
        lifetimePointsEarned,
        pointsToNextTier,
        nextTierName,
      ];
}

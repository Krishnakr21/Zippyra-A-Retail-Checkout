import 'package:equatable/equatable.dart';

class TierInfo extends Equatable {
  final String tier;
  final String displayName;
  final int minLifetimePoints;
  final double earnMultiplier;

  const TierInfo({
    required this.tier,
    required this.displayName,
    required this.minLifetimePoints,
    required this.earnMultiplier,
  });

  @override
  List<Object?> get props => [tier, displayName, minLifetimePoints, earnMultiplier];
}

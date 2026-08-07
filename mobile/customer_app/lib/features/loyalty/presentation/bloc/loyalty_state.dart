part of 'loyalty_bloc.dart';

abstract class LoyaltyState extends Equatable {
  const LoyaltyState();

  @override
  List<Object?> get props => [];
}

class LoyaltyInitial extends LoyaltyState {}

class LoyaltyLoading extends LoyaltyState {}

class LoyaltyLoaded extends LoyaltyState {
  final int pointsBalance;
  final int pointsReserved;
  final String tier;
  final String tierDisplayName;
  final int lifetimePointsEarned;
  final int? pointsToNextTier;
  final String? nextTierName;

  const LoyaltyLoaded({
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

class LoyaltyError extends LoyaltyState {
  final String message;

  const LoyaltyError(this.message);

  @override
  List<Object?> get props => [message];
}

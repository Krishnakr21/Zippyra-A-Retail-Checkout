import 'package:equatable/equatable.dart';

class SplitEstimate extends Equatable {
  final int originalTotalPaise;
  final int loyaltyDiscountPaise;
  final int payableAmountPaise;
  final int pointsBalance;

  const SplitEstimate({
    required this.originalTotalPaise,
    required this.loyaltyDiscountPaise,
    required this.payableAmountPaise,
    required this.pointsBalance,
  });

  @override
  List<Object?> get props => [
        originalTotalPaise,
        loyaltyDiscountPaise,
        payableAmountPaise,
        pointsBalance,
      ];
}

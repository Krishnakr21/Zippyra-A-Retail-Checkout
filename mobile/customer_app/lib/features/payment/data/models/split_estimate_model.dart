import '../../domain/entities/split_estimate.dart';

class SplitEstimateModel extends SplitEstimate {
  const SplitEstimateModel({
    required super.originalTotalPaise,
    required super.loyaltyDiscountPaise,
    required super.payableAmountPaise,
    required super.pointsBalance,
  });

  factory SplitEstimateModel.fromJson(Map<String, dynamic> json) {
    return SplitEstimateModel(
      originalTotalPaise: (json['original_total_paise'] ?? 0) as int,
      loyaltyDiscountPaise: (json['loyalty_discount_paise'] ?? 0) as int,
      payableAmountPaise: (json['payable_amount_paise'] ?? 0) as int,
      pointsBalance: (json['points_balance'] ?? 0) as int,
    );
  }
}

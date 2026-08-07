import '../../domain/entities/payment_intent.dart';

class PaymentIntentModel extends PaymentIntent {
  const PaymentIntentModel({
    required super.paymentId,
    required super.gateway,
    required super.gatewayOrderId,
    required super.gatewayKeyId,
    required super.payableAmountPaise,
    required super.expiresAt,
  });

  factory PaymentIntentModel.fromJson(Map<String, dynamic> json) {
    final expiresAtStr = json['expires_at'] as String?;
    final expiresAt = expiresAtStr != null ? DateTime.parse(expiresAtStr) : DateTime.now().add(const Duration(minutes: 10));

    return PaymentIntentModel(
      paymentId: json['payment_id'] as String? ?? '',
      gateway: json['gateway'] as String? ?? 'razorpay',
      gatewayOrderId: json['gateway_order_id'] as String? ?? '',
      gatewayKeyId: json['gateway_key_id'] as String? ?? '',
      payableAmountPaise: (json['payable_amount_paise'] ?? 0) as int,
      expiresAt: expiresAt,
    );
  }
}

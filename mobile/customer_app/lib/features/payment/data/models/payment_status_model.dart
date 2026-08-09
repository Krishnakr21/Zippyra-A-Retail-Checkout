import '../../domain/entities/payment_status_entity.dart';

class PaymentStatusModel extends PaymentStatusEntity {
  const PaymentStatusModel({
    required super.status,
    super.paymentMethod,
    super.gatewayPaymentId,
    super.failureReason,
  });

  factory PaymentStatusModel.fromJson(Map<String, dynamic> json) {
    return PaymentStatusModel(
      status: json['status'] as String? ?? 'PENDING',
      paymentMethod: json['payment_method'] as String?,
      gatewayPaymentId: json['gateway_payment_id'] as String?,
      failureReason: json['failure_reason'] as String?,
    );
  }
}

import 'package:equatable/equatable.dart';

class PaymentIntent extends Equatable {
  final String paymentId;
  final String gateway;
  final String gatewayOrderId;
  final String gatewayKeyId;
  final int payableAmountPaise;
  final DateTime expiresAt;

  const PaymentIntent({
    required this.paymentId,
    required this.gateway,
    required this.gatewayOrderId,
    required this.gatewayKeyId,
    required this.payableAmountPaise,
    required this.expiresAt,
  });

  @override
  List<Object?> get props => [
        paymentId,
        gateway,
        gatewayOrderId,
        gatewayKeyId,
        payableAmountPaise,
        expiresAt,
      ];
}

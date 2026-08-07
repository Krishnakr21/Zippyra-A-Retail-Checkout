import 'package:equatable/equatable.dart';

class PaymentStatusEntity extends Equatable {
  final String status;
  final String? paymentMethod;
  final String? gatewayPaymentId;
  final String? failureReason;

  const PaymentStatusEntity({
    required this.status,
    this.paymentMethod,
    this.gatewayPaymentId,
    this.failureReason,
  });

  bool get isCaptured => status == 'CAPTURED';
  bool get isFailed => status == 'FAILED';
  bool get isPending => status == 'PENDING' || status == 'INITIATED' || status == 'AUTHORIZED';

  @override
  List<Object?> get props => [status, paymentMethod, gatewayPaymentId, failureReason];
}

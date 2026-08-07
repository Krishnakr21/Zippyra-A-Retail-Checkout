part of 'payment_bloc.dart';

abstract class PaymentState extends Equatable {
  const PaymentState();

  @override
  List<Object?> get props => [];
}

class PaymentInitial extends PaymentState {}

class PaymentRecoveryChecking extends PaymentState {}

class PaymentCheckoutReady extends PaymentState {
  final int totalPaise;
  final int estimatedPayablePaise;
  final String selectedMethod;
  final int pointsToRedeem;
  final int pointsBalance;

  const PaymentCheckoutReady({
    required this.totalPaise,
    required this.estimatedPayablePaise,
    required this.selectedMethod,
    this.pointsToRedeem = 0,
    this.pointsBalance = 500,
  });

  PaymentCheckoutReady copyWith({
    int? totalPaise,
    int? estimatedPayablePaise,
    String? selectedMethod,
    int? pointsToRedeem,
    int? pointsBalance,
  }) {
    return PaymentCheckoutReady(
      totalPaise: totalPaise ?? this.totalPaise,
      estimatedPayablePaise: estimatedPayablePaise ?? this.estimatedPayablePaise,
      selectedMethod: selectedMethod ?? this.selectedMethod,
      pointsToRedeem: pointsToRedeem ?? this.pointsToRedeem,
      pointsBalance: pointsBalance ?? this.pointsBalance,
    );
  }

  @override
  List<Object?> get props => [
        totalPaise,
        estimatedPayablePaise,
        selectedMethod,
        pointsToRedeem,
        pointsBalance,
      ];
}

class PaymentEstimating extends PaymentState {}

class PaymentInitiating extends PaymentState {}

class PaymentProcessing extends PaymentState {
  final String paymentId;
  final int attemptsRemaining;

  const PaymentProcessing({
    required this.paymentId,
    required this.attemptsRemaining,
  });

  @override
  List<Object?> get props => [paymentId, attemptsRemaining];
}

class PaymentSuccess extends PaymentState {
  final String paymentId;

  const PaymentSuccess(this.paymentId);

  @override
  List<Object?> get props => [paymentId];
}

class PaymentFailed extends PaymentState {
  final String reason;

  const PaymentFailed(this.reason);

  @override
  List<Object?> get props => [reason];
}

class PaymentPendingTimeout extends PaymentState {}

class PaymentError extends PaymentState {
  final String errorCode;
  final String message;

  const PaymentError({
    required this.errorCode,
    required this.message,
  });

  @override
  List<Object?> get props => [errorCode, message];
}

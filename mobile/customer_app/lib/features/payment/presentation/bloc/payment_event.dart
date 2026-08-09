part of 'payment_bloc.dart';

abstract class PaymentEvent extends Equatable {
  const PaymentEvent();

  @override
  List<Object?> get props => [];
}

class LoyaltyPointsSliderChanged extends PaymentEvent {
  final int points;
  final String checkoutSessionId;

  const LoyaltyPointsSliderChanged({
    required this.points,
    required this.checkoutSessionId,
  });

  @override
  List<Object?> get props => [points, checkoutSessionId];
}

class PaymentMethodSelected extends PaymentEvent {
  final String method;

  const PaymentMethodSelected(this.method);

  @override
  List<Object?> get props => [method];
}

class PaymentInitiateRequested extends PaymentEvent {
  final String checkoutSessionId;

  const PaymentInitiateRequested(this.checkoutSessionId);

  @override
  List<Object?> get props => [checkoutSessionId];
}

class PaymentStatusPollTicked extends PaymentEvent {
  final String paymentId;

  const PaymentStatusPollTicked(this.paymentId);

  @override
  List<Object?> get props => [paymentId];
}

class PaymentDeepLinkReceived extends PaymentEvent {
  final String paymentId;

  const PaymentDeepLinkReceived(this.paymentId);

  @override
  List<Object?> get props => [paymentId];
}

class PendingPaymentCheckRequested extends PaymentEvent {}

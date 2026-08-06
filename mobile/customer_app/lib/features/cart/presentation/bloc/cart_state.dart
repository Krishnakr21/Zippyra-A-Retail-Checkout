part of 'cart_bloc.dart';

abstract class CartState extends Equatable {
  const CartState();

  @override
  List<Object?> get props => [];
}

class CartInitial extends CartState {}

class CartLoading extends CartState {}

class CartEmpty extends CartState {}

class CartLoaded extends CartState {
  final CartSummary summary;

  const CartLoaded(this.summary);

  @override
  List<Object?> get props => [summary];
}

class CartItemOutOfStock extends CartState {
  final String barcode;
  final String message;
  final CartSummary? previousSummary;

  const CartItemOutOfStock({
    required this.barcode,
    required this.message,
    this.previousSummary,
  });

  @override
  List<Object?> get props => [barcode, message, previousSummary];
}

class CartPriceChanged extends CartState {
  final CartSummary? updatedCart;
  final String message;

  const CartPriceChanged({
    this.updatedCart,
    required this.message,
  });

  @override
  List<Object?> get props => [updatedCart, message];
}

class CartLocked extends CartState {
  final String message;

  const CartLocked(this.message);

  @override
  List<Object?> get props => [message];
}

class CartCouponError extends CartState {
  final String errorCode;
  final String message;
  final CartSummary? summary;

  const CartCouponError({
    required this.errorCode,
    required this.message,
    this.summary,
  });

  @override
  List<Object?> get props => [errorCode, message, summary];
}

class CartCheckoutReady extends CartState {
  final String checkoutSessionId;
  final int totalPaise;
  final DateTime expiresAt;

  const CartCheckoutReady({
    required this.checkoutSessionId,
    required this.totalPaise,
    required this.expiresAt,
  });

  @override
  List<Object?> get props => [checkoutSessionId, totalPaise, expiresAt];
}

class CartError extends CartState {
  final String errorCode;
  final String message;

  const CartError({
    required this.errorCode,
    required this.message,
  });

  @override
  List<Object?> get props => [errorCode, message];
}

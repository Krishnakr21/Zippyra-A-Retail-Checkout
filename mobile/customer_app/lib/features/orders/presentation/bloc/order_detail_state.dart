part of 'order_detail_bloc.dart';

abstract class OrderDetailState extends Equatable {
  const OrderDetailState();

  @override
  List<Object?> get props => [];
}

class OrderDetailInitial extends OrderDetailState {}

class OrderDetailLoading extends OrderDetailState {}

class OrderDetailLoaded extends OrderDetailState {
  final OrderDetail order;

  const OrderDetailLoaded(this.order);

  @override
  List<Object?> get props => [order];
}

class OrderDetailError extends OrderDetailState {
  final String message;

  const OrderDetailError(this.message);

  @override
  List<Object?> get props => [message];
}

class ReturnSubmitting extends OrderDetailState {}

class ReturnSubmitted extends OrderDetailState {
  final String orderId;

  const ReturnSubmitted(this.orderId);

  @override
  List<Object?> get props => [orderId];
}

class ReturnFailed extends OrderDetailState {
  final String errorCode;
  final String message;
  final OrderDetail? currentOrder;

  const ReturnFailed({
    required this.errorCode,
    required this.message,
    this.currentOrder,
  });

  @override
  List<Object?> get props => [errorCode, message, currentOrder];
}

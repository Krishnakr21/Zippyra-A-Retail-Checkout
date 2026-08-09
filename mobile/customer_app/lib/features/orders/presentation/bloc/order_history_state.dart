part of 'order_history_bloc.dart';

abstract class OrderHistoryState extends Equatable {
  const OrderHistoryState();

  @override
  List<Object?> get props => [];
}

class OrderHistoryInitial extends OrderHistoryState {}

class OrderHistoryLoading extends OrderHistoryState {}

class OrderHistoryLoaded extends OrderHistoryState {
  final List<OrderSummary> orders;
  final bool hasMore;

  const OrderHistoryLoaded({
    required this.orders,
    required this.hasMore,
  });

  @override
  List<Object?> get props => [orders, hasMore];
}

class OrderHistoryError extends OrderHistoryState {
  final String errorCode;

  const OrderHistoryError(this.errorCode);

  @override
  List<Object?> get props => [errorCode];
}

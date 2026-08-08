part of 'order_exit_cubit.dart';

abstract class OrderExitState extends Equatable {
  const OrderExitState();

  @override
  List<Object?> get props => [];
}

class OrderExitInitial extends OrderExitState {}

class OrderExitLoading extends OrderExitState {}

class OrderExitLoaded extends OrderExitState {
  final ExitToken exitToken;

  const OrderExitLoaded(this.exitToken);

  @override
  List<Object?> get props => [exitToken];
}

class OrderExitError extends OrderExitState {
  final String message;

  const OrderExitError(this.message);

  @override
  List<Object?> get props => [message];
}

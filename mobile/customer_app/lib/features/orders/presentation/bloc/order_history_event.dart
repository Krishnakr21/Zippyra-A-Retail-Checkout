part of 'order_history_bloc.dart';

abstract class OrderHistoryEvent extends Equatable {
  const OrderHistoryEvent();

  @override
  List<Object?> get props => [];
}

class OrderHistoryRequested extends OrderHistoryEvent {
  final bool refresh;

  const OrderHistoryRequested({this.refresh = false});

  @override
  List<Object?> get props => [refresh];
}

class OrderHistoryNextPageRequested extends OrderHistoryEvent {}

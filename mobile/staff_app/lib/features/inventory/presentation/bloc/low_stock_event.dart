part of 'low_stock_bloc.dart';

abstract class LowStockEvent extends Equatable {
  const LowStockEvent();

  @override
  List<Object?> get props => [];
}

class LowStockRequested extends LowStockEvent {
  final String storeId;

  const LowStockRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

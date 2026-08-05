part of 'low_stock_bloc.dart';

abstract class LowStockState extends Equatable {
  const LowStockState();

  @override
  List<Object?> get props => [];
}

class LowStockInitial extends LowStockState {}

class LowStockLoading extends LowStockState {}

class LowStockLoaded extends LowStockState {
  final List<LowStockItem> items;

  const LowStockLoaded(this.items);

  @override
  List<Object?> get props => [items];
}

class LowStockError extends LowStockState {
  final String message;

  const LowStockError(this.message);

  @override
  List<Object?> get props => [message];
}

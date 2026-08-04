part of 'stock_count_bloc.dart';

abstract class StockCountEvent extends Equatable {
  const StockCountEvent();

  @override
  List<Object?> get props => [];
}

class ItemScanned extends StockCountEvent {
  final String barcode;
  final String? name;

  const ItemScanned({required this.barcode, this.name});

  @override
  List<Object?> get props => [barcode, name];
}

class ItemCountEdited extends StockCountEvent {
  final String barcode;
  final int qty;

  const ItemCountEdited({required this.barcode, required this.qty});

  @override
  List<Object?> get props => [barcode, qty];
}

class CountSubmitted extends StockCountEvent {
  final String storeId;

  const CountSubmitted(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

import 'package:equatable/equatable.dart';

class StockCountEntry extends Equatable {
  final String barcode;
  final String name;
  final int countedQty;

  const StockCountEntry({
    required this.barcode,
    required this.name,
    required this.countedQty,
  });

  StockCountEntry copyWith({int? countedQty}) {
    return StockCountEntry(
      barcode: barcode,
      name: name,
      countedQty: countedQty ?? this.countedQty,
    );
  }

  @override
  List<Object?> get props => [barcode, name, countedQty];
}

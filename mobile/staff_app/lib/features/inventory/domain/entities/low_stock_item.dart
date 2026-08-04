import 'package:equatable/equatable.dart';

class LowStockItem extends Equatable {
  final String barcode;
  final String productName;
  final int onHandQty;
  final int reorderPoint;
  final int reorderQty;

  const LowStockItem({
    required this.barcode,
    required this.productName,
    required this.onHandQty,
    required this.reorderPoint,
    required this.reorderQty,
  });

  @override
  List<Object?> get props => [barcode, productName, onHandQty, reorderPoint, reorderQty];
}

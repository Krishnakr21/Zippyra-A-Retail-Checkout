import 'package:equatable/equatable.dart';

class ReturnableItem extends Equatable {
  final String barcode;
  final String name;
  final int qty;
  final int returnedQty;
  final bool isReturnable;
  final int pricePaise;

  const ReturnableItem({
    required this.barcode,
    required this.name,
    required this.qty,
    this.returnedQty = 0,
    required this.isReturnable,
    required this.pricePaise,
  });

  int get remainingReturnableQty => isReturnable ? (qty - returnedQty).clamp(0, qty) : 0;

  @override
  List<Object?> get props => [barcode, name, qty, returnedQty, isReturnable, pricePaise];
}

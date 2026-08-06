import 'package:equatable/equatable.dart';

class CartItem extends Equatable {
  final String barcode;
  final String name;
  final int qty;
  final int pricePaise;
  final int lineTotalPaise;
  final String? imageUrl;

  const CartItem({
    required this.barcode,
    required this.name,
    required this.qty,
    required this.pricePaise,
    required this.lineTotalPaise,
    this.imageUrl,
  });

  @override
  List<Object?> get props => [barcode, name, qty, pricePaise, lineTotalPaise, imageUrl];
}

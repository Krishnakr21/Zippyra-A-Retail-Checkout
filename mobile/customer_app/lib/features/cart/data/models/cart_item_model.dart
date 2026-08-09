import '../../domain/entities/cart_item.dart';

class CartItemModel extends CartItem {
  const CartItemModel({
    required super.barcode,
    required super.name,
    required super.qty,
    required super.pricePaise,
    required super.lineTotalPaise,
    super.imageUrl,
  });

  factory CartItemModel.fromJson(Map<String, dynamic> json) {
    final pricePaise = (json['price_paise_snapshot'] ?? json['price_paise'] ?? 0) as int;
    final qty = (json['qty'] ?? 1) as int;
    final lineTotal = (json['line_total_paise'] ?? (pricePaise * qty)) as int;

    return CartItemModel(
      barcode: json['barcode'] as String? ?? '',
      name: json['name'] as String? ?? 'Product',
      qty: qty,
      pricePaise: pricePaise,
      lineTotalPaise: lineTotal,
      imageUrl: json['image_url'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'barcode': barcode,
      'name': name,
      'qty': qty,
      'price_paise': pricePaise,
      'line_total_paise': lineTotalPaise,
      if (imageUrl != null) 'image_url': imageUrl,
    };
  }
}

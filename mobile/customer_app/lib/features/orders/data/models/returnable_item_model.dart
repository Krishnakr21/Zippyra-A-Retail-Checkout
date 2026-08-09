import '../../domain/entities/returnable_item.dart';

class ReturnableItemModel extends ReturnableItem {
  const ReturnableItemModel({
    required super.barcode,
    required super.name,
    required super.qty,
    super.returnedQty = 0,
    required super.isReturnable,
    required super.pricePaise,
  });

  factory ReturnableItemModel.fromJson(Map<String, dynamic> json) {
    return ReturnableItemModel(
      barcode: json['barcode'] as String? ?? '',
      name: json['name'] as String? ?? 'Item',
      qty: (json['qty'] ?? 1) as int,
      returnedQty: (json['returned_qty'] ?? 0) as int,
      isReturnable: (json['is_returnable'] ?? true) as bool,
      pricePaise: (json['price_paise'] ?? 0) as int,
    );
  }
}

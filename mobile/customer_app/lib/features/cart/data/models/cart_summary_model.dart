import '../../domain/entities/cart_item.dart';
import '../../domain/entities/cart_summary.dart';
import 'cart_item_model.dart';

class CartSummaryModel extends CartSummary {
  const CartSummaryModel({
    required List<CartItem> items,
    required super.subtotalPaise,
    required super.discountPaise,
    required super.cgstPaise,
    required super.sgstPaise,
    required super.igstPaise,
    required super.totalPaise,
    required super.appliedOffers,
    super.couponCode,
    required super.itemCount,
  }) : super(items: items);

  factory CartSummaryModel.fromJson(Map<String, dynamic> json) {
    final rawItems = json['items'] as List<dynamic>? ?? [];
    final items = rawItems.map((e) => CartItemModel.fromJson(e as Map<String, dynamic>)).toList();
    final rawOffers = json['applied_offers'] as List<dynamic>? ?? [];
    final appliedOffers = rawOffers.map((e) => e.toString()).toList();

    return CartSummaryModel(
      items: items,
      subtotalPaise: (json['subtotal_paise'] ?? 0) as int,
      discountPaise: (json['discount_paise'] ?? 0) as int,
      cgstPaise: (json['cgst_paise'] ?? 0) as int,
      sgstPaise: (json['sgst_paise'] ?? 0) as int,
      igstPaise: (json['igst_paise'] ?? 0) as int,
      totalPaise: (json['total_paise'] ?? 0) as int,
      appliedOffers: appliedOffers,
      couponCode: json['coupon_code'] as String?,
      itemCount: (json['item_count'] ?? 0) as int,
    );
  }
}

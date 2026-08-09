import 'package:equatable/equatable.dart';
import 'cart_item.dart';

class CartSummary extends Equatable {
  final List<CartItem> items;
  final int subtotalPaise;
  final int discountPaise;
  final int cgstPaise;
  final int sgstPaise;
  final int igstPaise;
  final int totalPaise;
  final List<String> appliedOffers;
  final String? couponCode;
  final int itemCount;

  const CartSummary({
    required this.items,
    required this.subtotalPaise,
    required this.discountPaise,
    required this.cgstPaise,
    required this.sgstPaise,
    required this.igstPaise,
    required this.totalPaise,
    required this.appliedOffers,
    this.couponCode,
    required this.itemCount,
  });

  @override
  List<Object?> get props => [
        items,
        subtotalPaise,
        discountPaise,
        cgstPaise,
        sgstPaise,
        igstPaise,
        totalPaise,
        appliedOffers,
        couponCode,
        itemCount,
      ];
}

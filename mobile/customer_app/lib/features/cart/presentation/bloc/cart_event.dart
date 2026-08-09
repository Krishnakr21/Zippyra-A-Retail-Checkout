part of 'cart_bloc.dart';

abstract class CartEvent extends Equatable {
  const CartEvent();

  @override
  List<Object?> get props => [];
}

class CartRefreshRequested extends CartEvent {
  final String storeId;
  const CartRefreshRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class ItemScanned extends CartEvent {
  final String storeId;
  final String barcode;
  final int qty;

  const ItemScanned({
    required this.storeId,
    required this.barcode,
    this.qty = 1,
  });

  @override
  List<Object?> get props => [storeId, barcode, qty];
}

class ItemQuantityChanged extends CartEvent {
  final String storeId;
  final String barcode;
  final int qty;

  const ItemQuantityChanged({
    required this.storeId,
    required this.barcode,
    required this.qty,
  });

  @override
  List<Object?> get props => [storeId, barcode, qty];
}

class ItemRemoved extends CartEvent {
  final String storeId;
  final String barcode;

  const ItemRemoved({
    required this.storeId,
    required this.barcode,
  });

  @override
  List<Object?> get props => [storeId, barcode];
}

class CartCleared extends CartEvent {
  final String? storeId;
  const CartCleared({this.storeId});

  @override
  List<Object?> get props => [storeId];
}

class CouponApplyRequested extends CartEvent {
  final String storeId;
  final String code;

  const CouponApplyRequested({
    required this.storeId,
    required this.code,
  });

  @override
  List<Object?> get props => [storeId, code];
}

class CouponRemoveRequested extends CartEvent {
  final String storeId;

  const CouponRemoveRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class CheckoutRequested extends CartEvent {
  final String storeId;

  const CheckoutRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

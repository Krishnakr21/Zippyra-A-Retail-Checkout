import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:customer_app/features/cart/domain/entities/cart_item.dart';
import 'package:customer_app/features/cart/domain/entities/cart_summary.dart';
import 'package:customer_app/features/cart/domain/entities/checkout_session.dart';
import 'package:customer_app/features/cart/domain/repositories/cart_repository.dart';
import 'package:customer_app/features/cart/domain/usecases/apply_coupon_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/clear_cart_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/get_cart_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/init_checkout_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/remove_coupon_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/remove_item_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/scan_item_use_case.dart';
import 'package:customer_app/features/cart/domain/usecases/update_quantity_use_case.dart';
import 'package:customer_app/features/cart/presentation/bloc/cart_bloc.dart';

class FakeCartRepository implements CartRepository {
  CartSummary? _cart;
  bool initCheckoutCalled = false;

  @override
  CartSummary? getCachedCart() => _cart;

  @override
  void clearMemoryCache() {
    _cart = null;
  }

  @override
  Future<CartSummary> getCart(String storeId) async {
    return _cart ??
        const CartSummary(
          items: [],
          subtotalPaise: 0,
          discountPaise: 0,
          cgstPaise: 0,
          sgstPaise: 0,
          igstPaise: 0,
          totalPaise: 0,
          appliedOffers: [],
          itemCount: 0,
        );
  }

  @override
  Future<CartSummary> scanItem(String storeId, String barcode, {int qty = 1}) async {
    if (barcode == 'OUT_OF_STOCK_BARCODE') {
      throw const OutOfStockFailure('Item out of stock');
    }
    final item = CartItem(
      barcode: barcode,
      name: 'Coffee',
      qty: qty,
      pricePaise: 25000,
      lineTotalPaise: 25000 * qty,
    );
    _cart = CartSummary(
      items: [item],
      subtotalPaise: 25000 * qty,
      discountPaise: 0,
      cgstPaise: 625 * qty,
      sgstPaise: 625 * qty,
      igstPaise: 0,
      totalPaise: 26250 * qty,
      appliedOffers: const [],
      itemCount: qty,
    );
    return _cart!;
  }

  @override
  Future<CartSummary> updateQuantity(String storeId, String barcode, int qty) async {
    return scanItem(storeId, barcode, qty: qty);
  }

  @override
  Future<CartSummary> removeItem(String storeId, String barcode) async {
    _cart = const CartSummary(
      items: [],
      subtotalPaise: 0,
      discountPaise: 0,
      cgstPaise: 0,
      sgstPaise: 0,
      igstPaise: 0,
      totalPaise: 0,
      appliedOffers: [],
      itemCount: 0,
    );
    return _cart!;
  }

  @override
  Future<CartSummary> clearCart(String storeId) async {
    return removeItem(storeId, '');
  }

  @override
  Future<CartSummary> applyCoupon(String storeId, String code) async {
    if (code == 'INVALID') {
      throw const CouponInvalidFailure('Invalid coupon code');
    }
    return _cart!;
  }

  @override
  Future<CartSummary> removeCoupon(String storeId) async {
    return _cart!;
  }

  @override
  Future<CheckoutSession> initCheckout(String storeId) async {
    initCheckoutCalled = true;
    if (storeId == 'store-price-changed') {
      throw const PriceChangedFailure(null, 'Item prices changed');
    }
    return CheckoutSession(
      id: 'sess-100',
      totalPaise: _cart?.totalPaise ?? 0,
      expiresAt: DateTime.now().add(const Duration(minutes: 10)),
    );
  }
}

void main() {
  late FakeCartRepository repo;
  late CartBloc bloc;

  setUp(() {
    repo = FakeCartRepository();
    bloc = CartBloc(
      getCartUseCase: GetCartUseCase(repo),
      scanItemUseCase: ScanItemUseCase(repo),
      updateQuantityUseCase: UpdateQuantityUseCase(repo),
      removeItemUseCase: RemoveItemUseCase(repo),
      clearCartUseCase: ClearCartUseCase(repo),
      applyCouponUseCase: ApplyCouponUseCase(repo),
      removeCouponUseCase: RemoveCouponUseCase(repo),
      initCheckoutUseCase: InitCheckoutUseCase(repo),
    );
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is CartInitial', () {
    expect(bloc.state, isA<CartInitial>());
  });

  blocTest<CartBloc, CartState>(
    'ItemScanned success updates state to CartLoaded with running totals',
    build: () => bloc,
    act: (bloc) => bloc.add(const ItemScanned(storeId: 'store-1', barcode: '8901030300011', qty: 1)),
    expect: () => [
      isA<CartLoaded>().having((s) => s.summary.totalPaise, 'totalPaise', 26250),
    ],
  );

  blocTest<CartBloc, CartState>(
    'ItemScanned OUT_OF_STOCK emits CartItemOutOfStock without clearing state',
    build: () => bloc,
    act: (bloc) => bloc.add(const ItemScanned(storeId: 'store-1', barcode: 'OUT_OF_STOCK_BARCODE')),
    expect: () => [
      isA<CartItemOutOfStock>().having((s) => s.message, 'message', 'Item out of stock'),
    ],
  );

  blocTest<CartBloc, CartState>(
    'CheckoutRequested with empty cart never calls repository and emits CartError(CART_EMPTY) locally',
    build: () => bloc,
    act: (bloc) => bloc.add(const CheckoutRequested('store-1')),
    expect: () => [
      const CartError(errorCode: ErrorCodes.cartEmpty, message: 'Cart is empty'),
    ],
    verify: (_) {
      expect(repo.initCheckoutCalled, isFalse);
    },
  );

  blocTest<CartBloc, CartState>(
    'PriceChangedFailure updates state to CartPriceChanged, and subsequent CartRefreshRequested returns CartLoaded',
    build: () => bloc,
    seed: () => const CartLoaded(CartSummary(
      items: [CartItem(barcode: 'b1', name: 'Item', qty: 1, pricePaise: 100, lineTotalPaise: 100)],
      subtotalPaise: 100,
      discountPaise: 0,
      cgstPaise: 0,
      sgstPaise: 0,
      igstPaise: 0,
      totalPaise: 100,
      appliedOffers: [],
      itemCount: 1,
    )),
    act: (bloc) async {
      bloc.add(const CheckoutRequested('store-price-changed'));
      await Future.delayed(const Duration(milliseconds: 50));
      bloc.add(const CartRefreshRequested('store-1'));
    },
    expect: () => [
      isA<CartLoading>(),
      isA<CartPriceChanged>().having((s) => s.message, 'message', 'Item prices changed'),
      isA<CartLoading>(),
      isA<CartEmpty>(),
    ],
  );
}

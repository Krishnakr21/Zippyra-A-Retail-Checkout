import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';

import '../../domain/entities/cart_summary.dart';
import '../../domain/usecases/apply_coupon_use_case.dart';
import '../../domain/usecases/clear_cart_use_case.dart';
import '../../domain/usecases/get_cart_use_case.dart';
import '../../domain/usecases/init_checkout_use_case.dart';
import '../../domain/usecases/remove_coupon_use_case.dart';
import '../../domain/usecases/remove_item_use_case.dart';
import '../../domain/usecases/scan_item_use_case.dart';
import '../../domain/usecases/update_quantity_use_case.dart';

part 'cart_event.dart';
part 'cart_state.dart';

class CartBloc extends Bloc<CartEvent, CartState> {
  final GetCartUseCase getCartUseCase;
  final ScanItemUseCase scanItemUseCase;
  final UpdateQuantityUseCase updateQuantityUseCase;
  final RemoveItemUseCase removeItemUseCase;
  final ClearCartUseCase clearCartUseCase;
  final ApplyCouponUseCase applyCouponUseCase;
  final RemoveCouponUseCase removeCouponUseCase;
  final InitCheckoutUseCase initCheckoutUseCase;

  CartSummary? _currentSummary;

  CartBloc({
    required this.getCartUseCase,
    required this.scanItemUseCase,
    required this.updateQuantityUseCase,
    required this.removeItemUseCase,
    required this.clearCartUseCase,
    required this.applyCouponUseCase,
    required this.removeCouponUseCase,
    required this.initCheckoutUseCase,
  }) : super(CartInitial()) {
    on<CartRefreshRequested>(_onCartRefreshRequested);
    on<ItemScanned>(_onItemScanned);
    on<ItemQuantityChanged>(_onItemQuantityChanged);
    on<ItemRemoved>(_onItemRemoved);
    on<CartCleared>(_onCartCleared);
    on<CouponApplyRequested>(_onCouponApplyRequested);
    on<CouponRemoveRequested>(_onCouponRemoveRequested);
    on<CheckoutRequested>(_onCheckoutRequested);
  }

  Future<void> _onCartRefreshRequested(
    CartRefreshRequested event,
    Emitter<CartState> emit,
  ) async {
    emit(CartLoading());
    try {
      final summary = await getCartUseCase(event.storeId);
      _currentSummary = summary;
      if (summary.items.isEmpty) {
        emit(CartEmpty());
      } else {
        emit(CartLoaded(summary));
      }
    } catch (e) {
      _emitError(e, emit);
    }
  }

  Future<void> _onItemScanned(
    ItemScanned event,
    Emitter<CartState> emit,
  ) async {
    try {
      final summary = await scanItemUseCase(event.storeId, event.barcode, qty: event.qty);
      _currentSummary = summary;
      if (summary.items.isEmpty) {
        emit(CartEmpty());
      } else {
        emit(CartLoaded(summary));
      }
    } catch (e) {
      if (e is OutOfStockFailure) {
        emit(CartItemOutOfStock(
          barcode: event.barcode,
          message: e.message,
          previousSummary: _currentSummary,
        ));
        // Restore cart loaded state after out of stock notification
        if (_currentSummary != null && _currentSummary!.items.isNotEmpty) {
          emit(CartLoaded(_currentSummary!));
        }
      } else {
        _emitError(e, emit);
      }
    }
  }

  Future<void> _onItemQuantityChanged(
    ItemQuantityChanged event,
    Emitter<CartState> emit,
  ) async {
    if (event.qty <= 0) {
      add(ItemRemoved(storeId: event.storeId, barcode: event.barcode));
      return;
    }

    try {
      final summary = await updateQuantityUseCase(event.storeId, event.barcode, event.qty);
      _currentSummary = summary;
      if (summary.items.isEmpty) {
        emit(CartEmpty());
      } else {
        emit(CartLoaded(summary));
      }
    } catch (e) {
      if (e is OutOfStockFailure) {
        emit(CartItemOutOfStock(
          barcode: event.barcode,
          message: e.message,
          previousSummary: _currentSummary,
        ));
        if (_currentSummary != null && _currentSummary!.items.isNotEmpty) {
          emit(CartLoaded(_currentSummary!));
        }
      } else {
        _emitError(e, emit);
      }
    }
  }

  Future<void> _onItemRemoved(
    ItemRemoved event,
    Emitter<CartState> emit,
  ) async {
    try {
      final summary = await removeItemUseCase(event.storeId, event.barcode);
      _currentSummary = summary;
      if (summary.items.isEmpty) {
        emit(CartEmpty());
      } else {
        emit(CartLoaded(summary));
      }
    } catch (e) {
      _emitError(e, emit);
    }
  }

  Future<void> _onCartCleared(
    CartCleared event,
    Emitter<CartState> emit,
  ) async {
    if (event.storeId != null && event.storeId!.isNotEmpty) {
      try {
        await clearCartUseCase(event.storeId!);
      } catch (_) {}
    }
    _currentSummary = null;
    emit(CartEmpty());
  }

  Future<void> _onCouponApplyRequested(
    CouponApplyRequested event,
    Emitter<CartState> emit,
  ) async {
    try {
      final summary = await applyCouponUseCase(event.storeId, event.code);
      _currentSummary = summary;
      emit(CartLoaded(summary));
    } catch (e) {
      if (e is CouponInvalidFailure || e is CouponExpiredFailure || e is CouponMinNotMetFailure) {
        final failure = e as Failure;
        emit(CartCouponError(
          errorCode: failure.code ?? ErrorCodes.couponInvalid,
          message: failure.message,
          summary: _currentSummary,
        ));
        if (_currentSummary != null && _currentSummary!.items.isNotEmpty) {
          emit(CartLoaded(_currentSummary!));
        }
      } else {
        _emitError(e, emit);
      }
    }
  }

  Future<void> _onCouponRemoveRequested(
    CouponRemoveRequested event,
    Emitter<CartState> emit,
  ) async {
    try {
      final summary = await removeCouponUseCase(event.storeId);
      _currentSummary = summary;
      emit(CartLoaded(summary));
    } catch (e) {
      _emitError(e, emit);
    }
  }

  Future<void> _onCheckoutRequested(
    CheckoutRequested event,
    Emitter<CartState> emit,
  ) async {
    final summary = _currentSummary ?? (state is CartLoaded ? (state as CartLoaded).summary : null);
    // Guard clause: if cart is empty, do NOT call repository
    if (state is CartEmpty || summary == null || summary.items.isEmpty) {
      emit(const CartError(
        errorCode: ErrorCodes.cartEmpty,
        message: 'Cart is empty',
      ));
      return;
    }

    emit(CartLoading());
    try {
      final session = await initCheckoutUseCase(event.storeId);
      emit(CartCheckoutReady(
        checkoutSessionId: session.id,
        totalPaise: session.totalPaise,
        expiresAt: session.expiresAt,
      ));
    } catch (e) {
      if (e is PriceChangedFailure) {
        emit(CartPriceChanged(
          updatedCart: e.updatedCart is CartSummary ? e.updatedCart as CartSummary : _currentSummary,
          message: e.message,
        ));
      } else if (e is CartLockedFailure) {
        emit(CartLocked(e.message));
      } else {
        _emitError(e, emit);
      }
    }
  }

  void _emitError(Object e, Emitter<CartState> emit) {
    if (e is Failure) {
      emit(CartError(errorCode: e.code ?? ErrorCodes.invalidRequest, message: e.message));
    } else {
      emit(CartError(errorCode: ErrorCodes.invalidRequest, message: e.toString()));
    }
  }
}

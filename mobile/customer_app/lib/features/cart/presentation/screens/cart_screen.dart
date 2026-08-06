import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

import '../bloc/cart_bloc.dart';
import '../../domain/entities/cart_summary.dart';
import '../widgets/cart_item_tile.dart';
import '../widgets/cart_totals_summary.dart';
import '../widgets/coupon_input_field.dart';
import '../widgets/offer_banner.dart';
import '../widgets/sticky_checkout_bar.dart';
import 'cart_empty_state.dart';
import 'out_of_stock_snackbar.dart';
import 'price_changed_sheet.dart';

class CartScreen extends StatefulWidget {
  final String storeId;

  const CartScreen({super.key, this.storeId = 'store-1'});

  @override
  State<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends State<CartScreen> {
  @override
  void initState() {
    super.initState();
    context.read<CartBloc>().add(CartRefreshRequested(widget.storeId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Your Cart'),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_sweep_outlined),
            tooltip: 'Clear Cart',
            onPressed: () {
              context.read<CartBloc>().add(CartCleared(storeId: widget.storeId));
            },
          ),
        ],
      ),
      body: BlocConsumer<CartBloc, CartState>(
        listener: (context, state) {
          if (state is CartCheckoutReady) {
            context.push('/payment/checkout', extra: {
              'checkout_session_id': state.checkoutSessionId,
              'total_paise': state.totalPaise,
              'expires_at': state.expiresAt.toIso8601String(),
            });
          } else if (state is CartItemOutOfStock) {
            showOutOfStockSnackBar(context, state.message);
          } else if (state is CartPriceChanged) {
            showModalBottomSheet(
              context: context,
              isScrollControlled: true,
              builder: (ctx) => PriceChangedSheet(
                updatedCart: state.updatedCart,
                message: state.message,
                onRefresh: () {
                  context.read<CartBloc>().add(CartRefreshRequested(widget.storeId));
                },
              ),
            );
          } else if (state is CartCouponError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message), backgroundColor: Colors.red),
            );
          } else if (state is CartError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message), backgroundColor: Colors.red),
            );
          }
        },
        builder: (context, state) {
          if (state is CartLoading && state is! CartLoaded) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is CartEmpty) {
            return const CartEmptyState();
          }

          CartSummary? summary;
          bool isLocked = false;
          String? couponErrorMsg;

          if (state is CartLoaded) {
            summary = state.summary;
          } else if (state is CartLocked) {
            isLocked = true;
            summary = (context.read<CartBloc>().state as dynamic).summary as CartSummary?;
          } else if (state is CartCouponError) {
            couponErrorMsg = state.message;
            summary = state.summary;
          }

          if (summary == null || summary.items.isEmpty) {
            return const CartEmptyState();
          }

          return Column(
            children: [
              if (isLocked)
                Container(
                  width: double.infinity,
                  color: Colors.orange[100],
                  padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
                  child: const Row(
                    children: [
                      Icon(Icons.lock, color: Colors.orange, size: 20),
                      SizedBox(width: 8),
                      Text(
                        'Checkout in progress...',
                        style: TextStyle(color: Colors.orange, fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                ),

              Expanded(
                child: RefreshIndicator(
                  onRefresh: () async {
                    context.read<CartBloc>().add(CartRefreshRequested(widget.storeId));
                  },
                  child: ListView(
                    padding: const EdgeInsets.only(bottom: 16),
                    children: [
                      // Offer Banner
                      OfferBanner(offers: summary.appliedOffers),

                      // Cart Items List
                      ...summary.items.map(
                        (item) => CartItemTile(
                          item: item,
                          onQuantityChanged: (newQty) {
                            context.read<CartBloc>().add(ItemQuantityChanged(
                                  storeId: widget.storeId,
                                  barcode: item.barcode,
                                  qty: newQty,
                                ));
                          },
                          onRemove: () {
                            context.read<CartBloc>().add(ItemRemoved(
                                  storeId: widget.storeId,
                                  barcode: item.barcode,
                                ));
                          },
                        ),
                      ),

                      const SizedBox(height: 12),

                      // Coupon Input
                      CouponInputField(
                        appliedCoupon: summary.couponCode,
                        errorMessage: couponErrorMsg,
                        onApply: (code) {
                          context.read<CartBloc>().add(CouponApplyRequested(
                                storeId: widget.storeId,
                                code: code,
                              ));
                        },
                        onRemove: () {
                          context.read<CartBloc>().add(CouponRemoveRequested(widget.storeId));
                        },
                      ),

                      // Totals Summary Breakdown
                      CartTotalsSummary(summary: summary),
                    ],
                  ),
                ),
              ),

              // Sticky Checkout Bar
              StickyCheckoutBar(
                totalPaise: summary.totalPaise,
                isEnabled: !isLocked && summary.items.isNotEmpty,
                isLoading: state is CartLoading,
                onCheckout: () {
                  context.read<CartBloc>().add(CheckoutRequested(widget.storeId));
                },
              ),
            ],
          );
        },
      ),
    );
  }
}

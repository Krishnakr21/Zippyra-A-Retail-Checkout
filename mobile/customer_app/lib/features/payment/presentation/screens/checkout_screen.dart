import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

import '../../../cart/domain/entities/cart_summary.dart';
import '../../../cart/presentation/bloc/cart_bloc.dart';
import '../../../cart/presentation/widgets/cart_totals_summary.dart';
import '../bloc/payment_bloc.dart';

class CheckoutScreen extends StatefulWidget {
  final String checkoutSessionId;
  final int totalPaise;
  final DateTime expiresAt;

  CheckoutScreen({
    super.key,
    this.checkoutSessionId = 'sess-100',
    this.totalPaise = 50000,
    DateTime? expiresAt,
  }) : expiresAt = expiresAt ?? DateTime.now();

  @override
  State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  int _selectedPoints = 0;
  String _selectedMethod = 'UPI';

  @override
  void initState() {
    super.initState();
    // Initialize PaymentBloc into CheckoutReady state
    context.read<PaymentBloc>().emit(PaymentCheckoutReady(
          totalPaise: widget.totalPaise,
          estimatedPayablePaise: widget.totalPaise,
          selectedMethod: _selectedMethod,
        ));
  }

  @override
  Widget build(BuildContext context) {
    // Get last CartSummary from CartBloc state if available
    final cartState = context.watch<CartBloc>().state;
    CartSummary? summary;
    if (cartState is CartLoaded) {
      summary = cartState.summary;
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Checkout Payment'),
      ),
      body: BlocConsumer<PaymentBloc, PaymentState>(
        listener: (context, state) {
          if (state is PaymentProcessing) {
            context.push('/payment/processing', extra: {
              'payment_id': state.paymentId,
            });
          } else if (state is PaymentFailed) {
            context.push('/payment/failed', extra: {
              'reason': state.reason,
            });
          }
        },
        builder: (context, state) {
          int payablePaise = widget.totalPaise;
          int pointsBalance = 500;

          if (state is PaymentCheckoutReady) {
            payablePaise = state.estimatedPayablePaise;
            pointsBalance = state.pointsBalance;
          }

          return Column(
            children: [
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    // Order Summary Breakdown (reusing CartTotalsSummary if available)
                    if (summary != null)
                      CartTotalsSummary(summary: summary)
                    else
                      Card(
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text('Order Total', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                              Text(
                                CurrencyFormatter.formatPaise(widget.totalPaise),
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                              ),
                            ],
                          ),
                        ),
                      ),

                    const SizedBox(height: 16),

                    // Loyalty Points Redemption Slider
                    Card(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                const Row(
                                  children: [
                                    Icon(Icons.stars, color: Colors.amber),
                                    SizedBox(width: 8),
                                    Text('Use Loyalty Points', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                                  ],
                                ),
                                Text('$pointsBalance pts available', style: const TextStyle(color: Colors.grey, fontSize: 12)),
                              ],
                            ),
                            const SizedBox(height: 12),
                            Slider(
                              value: _selectedPoints.toDouble(),
                              min: 0,
                              max: pointsBalance.toDouble(),
                              divisions: (pointsBalance / 50).clamp(1, 20).toInt(),
                              activeColor: ZippyraColors.primary,
                              label: '$_selectedPoints pts',
                              onChanged: (val) {
                                setState(() {
                                  _selectedPoints = val.toInt();
                                });
                                context.read<PaymentBloc>().add(LoyaltyPointsSliderChanged(
                                      points: _selectedPoints,
                                      checkoutSessionId: widget.checkoutSessionId,
                                    ));
                              },
                            ),
                            if (_selectedPoints > 0)
                              Text(
                                'Redeeming $_selectedPoints pts (Saved ${CurrencyFormatter.formatPaise((_selectedPoints / 100).toInt() * 100)})',
                                style: const TextStyle(color: Colors.green, fontWeight: FontWeight.bold, fontSize: 13),
                              ),
                          ],
                        ),
                      ),
                    ),

                    const SizedBox(height: 16),

                    // Payment Method Selector
                    Card(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('Select Payment Method', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                            const SizedBox(height: 8),
                            RadioListTile<String>(
                              title: const Text('UPI (GPay / PhonePe / Paytm)'),
                              subtitle: const Text('Fastest checkout'),
                              value: 'UPI',
                              groupValue: _selectedMethod,
                              activeColor: ZippyraColors.primary,
                              onChanged: (val) {
                                if (val != null) {
                                  setState(() => _selectedMethod = val);
                                  context.read<PaymentBloc>().add(PaymentMethodSelected(val));
                                }
                              },
                            ),
                            RadioListTile<String>(
                              title: const Text('Credit / Debit Card'),
                              value: 'CARD',
                              groupValue: _selectedMethod,
                              activeColor: ZippyraColors.primary,
                              onChanged: (val) {
                                if (val != null) {
                                  setState(() => _selectedMethod = val);
                                  context.read<PaymentBloc>().add(PaymentMethodSelected(val));
                                }
                              },
                            ),
                            RadioListTile<String>(
                              title: const Text('Netbanking'),
                              value: 'NETBANKING',
                              groupValue: _selectedMethod,
                              activeColor: ZippyraColors.primary,
                              onChanged: (val) {
                                if (val != null) {
                                  setState(() => _selectedMethod = val);
                                  context.read<PaymentBloc>().add(PaymentMethodSelected(val));
                                }
                              },
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),

              // Bottom Pay Button
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  boxShadow: [
                    BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 10, offset: const Offset(0, -5)),
                  ],
                ),
                child: SafeArea(
                  child: SizedBox(
                    width: double.infinity,
                    height: 50,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: ZippyraColors.primary,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      onPressed: state is PaymentInitiating
                          ? null
                          : () {
                              context.read<PaymentBloc>().add(PaymentInitiateRequested(widget.checkoutSessionId));
                            },
                      child: state is PaymentInitiating
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Text(
                              'Pay ${CurrencyFormatter.formatPaise(payablePaise)}',
                              style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                            ),
                    ),
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

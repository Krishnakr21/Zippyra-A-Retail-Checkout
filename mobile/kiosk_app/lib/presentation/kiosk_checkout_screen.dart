import 'dart:async';
import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class KioskItem {
  final String sku;
  final String name;
  final double price;
  int quantity;

  KioskItem({required this.sku, required this.name, required this.price, this.quantity = 1});
}

class KioskCheckoutScreen extends StatefulWidget {
  final String storeId;
  final String customerPhone;
  final Function(String orderId, String exitQr) onPaymentCompleted;
  final VoidCallback onIdleTimeout;

  const KioskCheckoutScreen({
    super.key,
    required this.storeId,
    required this.customerPhone,
    required this.onPaymentCompleted,
    required this.onIdleTimeout,
  });

  @override
  State<KioskCheckoutScreen> createState() => _KioskCheckoutScreenState();
}

class _KioskCheckoutScreenState extends State<KioskCheckoutScreen> {
  final List<KioskItem> _cartItems = [
    KioskItem(sku: 'SKU-ORGANIC-MILK-1L', name: 'Organic Whole Milk 1L', price: 75.0, quantity: 2),
    KioskItem(sku: 'SKU-WHOLE-WHEAT-BREAD', name: 'Artisan Whole Wheat Bread', price: 45.0, quantity: 1),
  ];

  Timer? _idleTimer;
  bool _isProcessingPayment = false;
  String _couponCode = '';

  @override
  void initState() {
    super.initState();
    _resetIdleTimer();
  }

  void _resetIdleTimer() {
    _idleTimer?.cancel();
    // 2-minute idle timeout in unattended kiosk mode
    _idleTimer = Timer(const Duration(minutes: 2), () {
      widget.onIdleTimeout();
    });
  }

  @override
  void dispose() {
    _idleTimer?.cancel();
    super.dispose();
  }

  double get _subtotal => _cartItems.fold(0, (sum, item) => sum + (item.price * item.quantity));
  double get _discount => _couponCode == 'SAVE50' ? 50.0 : 0.0;
  double get _total => (_subtotal - _discount).clamp(0, 99999);

  void _addItem(String sku, String name, double price) {
    _resetIdleTimer();
    setState(() {
      final idx = _cartItems.indexWhere((it) => it.sku == sku);
      if (idx >= 0) {
        _cartItems[idx].quantity++;
      } else {
        _cartItems.add(KioskItem(sku: sku, name: name, price: price));
      }
    });
  }

  void _processPayment() {
    _resetIdleTimer();
    setState(() {
      _isProcessingPayment = true;
    });

    // Simulate Razorpay UPI QR payment processing
    Future.delayed(const Duration(seconds: 2), () {
      final orderId = 'ORD-KSK-${DateTime.now().millisecondsSinceEpoch}';
      final exitQr = 'EXIT-PASS-$orderId-${widget.storeId}';
      widget.onPaymentCompleted(orderId, exitQr);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Listener(
      onPointerDown: (_) => _resetIdleTimer(),
      child: Scaffold(
        backgroundColor: ZippyraColors.darkBackground,
        body: SafeArea(
          child: Column(
            children: [
              // Top Bar
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                color: ZippyraColors.darkCard,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.account_circle, color: ZippyraColors.primary, size: 28),
                        const SizedBox(width: 12),
                        Text(
                          'Customer: +91 ${widget.customerPhone}',
                          style: const TextStyle(fontSize: 18, color: Colors.white, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(backgroundColor: Colors.red.withOpacity(0.2)),
                      onPressed: widget.onIdleTimeout,
                      icon: const Icon(Icons.clear_all, color: Colors.redAccent),
                      label: const Text('Clear Cart & Exit', style: TextStyle(color: Colors.redAccent)),
                    ),
                  ],
                ),
              ),

              // Main Cart Content
              Expanded(
                child: Row(
                  children: [
                    // Left Column: Item List & Barcode Scanner Simulation
                    Expanded(
                      flex: 6,
                      child: Padding(
                        padding: const EdgeInsets.all(24.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Scanned Cart Items (${_cartItems.fold(0, (sum, i) => sum + i.quantity)})',
                              style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white),
                            ),
                            const SizedBox(height: 16),

                            Expanded(
                              child: ListView.builder(
                                itemCount: _cartItems.length,
                                itemBuilder: (context, idx) {
                                  final item = _cartItems[idx];
                                  return Card(
                                    color: ZippyraColors.darkCard,
                                    margin: const EdgeInsets.only(bottom: 12),
                                    child: ListTile(
                                      title: Text(item.name, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18)),
                                      subtitle: Text('SKU: ${item.sku} • ₹${item.price.toStringAsFixed(2)} each', style: const TextStyle(color: Colors.white70)),
                                      trailing: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          IconButton(
                                            icon: const Icon(Icons.remove_circle, color: Colors.redAccent, size: 28),
                                            onPressed: () {
                                              setState(() {
                                                if (item.quantity > 1) {
                                                  item.quantity--;
                                                } else {
                                                  _cartItems.removeAt(idx);
                                                }
                                              });
                                            },
                                          ),
                                          Text('${item.quantity}', style: const TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.bold)),
                                          IconButton(
                                            icon: const Icon(Icons.add_circle, color: ZippyraColors.primary, size: 28),
                                            onPressed: () {
                                              setState(() {
                                                item.quantity++;
                                              });
                                            },
                                          ),
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
                            ),

                            // Quick Demo Item Scan Buttons
                            Wrap(
                              spacing: 12,
                              children: [
                                ActionChip(
                                  backgroundColor: ZippyraColors.primary.withOpacity(0.2),
                                  label: const Text('+ Scan Olive Oil ₹350', style: TextStyle(color: ZippyraColors.primary)),
                                  onPressed: () => _addItem('SKU-OLIVE-OIL', 'Extra Virgin Olive Oil 500ml', 350.0),
                                ),
                                ActionChip(
                                  backgroundColor: ZippyraColors.primary.withOpacity(0.2),
                                  label: const Text('+ Apply Coupon SAVE50', style: TextStyle(color: ZippyraColors.primary)),
                                  onPressed: () {
                                    setState(() {
                                      _couponCode = 'SAVE50';
                                    });
                                  },
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),

                    // Right Column: Summary & UPI QR Payment
                    Expanded(
                      flex: 5,
                      child: Container(
                        color: ZippyraColors.darkCard,
                        padding: const EdgeInsets.all(32),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('Order Summary', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.white)),
                            const Divider(color: Colors.white24, height: 32),

                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                const Text('Subtotal', style: TextStyle(fontSize: 18, color: Colors.white70)),
                                Text('₹${_subtotal.toStringAsFixed(2)}', style: const TextStyle(fontSize: 18, color: Colors.white)),
                              ],
                            ),
                            if (_discount > 0)
                              Padding(
                                padding: const EdgeInsets.only(top: 8.0),
                                child: Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    Text('Discount ($_couponCode)', style: const TextStyle(fontSize: 18, color: ZippyraColors.primary)),
                                    Text('-₹${_discount.toStringAsFixed(2)}', style: const TextStyle(fontSize: 18, color: ZippyraColors.primary)),
                                  ],
                                ),
                              ),
                            const Spacer(),

                            // Total Amount Box
                            Container(
                              padding: const EdgeInsets.all(20),
                              decoration: BoxDecoration(
                                color: Colors.black45,
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: ZippyraColors.primary),
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  const Text('TOTAL PAYABLE', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.white)),
                                  Text('₹${_total.toStringAsFixed(2)}', style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: ZippyraColors.primary)),
                                ],
                              ),
                            ),
                            const SizedBox(height: 24),

                            if (_isProcessingPayment)
                              const Center(child: CircularProgressIndicator(color: ZippyraColors.primary))
                            else
                              SizedBox(
                                width: double.infinity,
                                height: 70,
                                child: ElevatedButton(
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: ZippyraColors.primary,
                                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                  ),
                                  onPressed: _cartItems.isEmpty ? null : _processPayment,
                                  child: const Text('SCAN UPI QR TO PAY', style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.black)),
                                ),
                              ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

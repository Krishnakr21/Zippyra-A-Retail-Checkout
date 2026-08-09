import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import '../../../injection_container.dart';
import '../../../core/services/root_detection_service.dart';

class CartScreen extends StatefulWidget {
  const CartScreen({super.key});

  @override
  State<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends State<CartScreen> {
  bool _isRooted = false;

  final List<Map<String, dynamic>> _cartItems = [
    {'emoji': '🧈', 'name': 'Amul Butter 500g', 'weight': '500g', 'price': 280, 'qty': 1},
    {'emoji': '🍞', 'name': 'Britannia Bread', 'weight': '400g', 'price': 45, 'qty': 1},
    {'emoji': '🥛', 'name': 'Amul Taza Milk 1L', 'weight': '1L', 'price': 62, 'qty': 1},
  ];

  @override
  void initState() {
    super.initState();
    _checkRoot();
  }

  Future<void> _checkRoot() async {
    final rootService = sl<CustomerRootDetectionService>();
    final rooted = await rootService.checkRootStatus();
    if (mounted) {
      setState(() => _isRooted = rooted);
    }
  }

  int get _totalPrice {
    int sum = 0;
    for (var item in _cartItems) {
      sum += (item['price'] as int) * (item['qty'] as int);
    }
    return sum;
  }

  int get _totalItems {
    int count = 0;
    for (var item in _cartItems) {
      count += (item['qty'] as int);
    }
    return count;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF4F5FA),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        title: Text(
          'Your Cart ($_totalItems Items)',
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: ZippyraColors.textPrimary),
        ),
      ),
      body: Column(
        children: [
          if (_isRooted) ...[
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              color: const Color(0xFFFFF0F0),
              child: const Row(
                children: [
                  Icon(Icons.warning_amber_rounded, color: ZippyraColors.errorRed, size: 20),
                  SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      'Rooted device detected. Express checkout security features are active.',
                      style: TextStyle(color: ZippyraColors.errorRed, fontSize: 11, fontWeight: FontWeight.bold),
                    ),
                  ),
                ],
              ),
            ),
          ],

          // Free Delivery & Exit Gate Ready Header Pill
          Container(
            margin: const EdgeInsets.all(14),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: const Color(0xFFE6FFF4),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFB3F0D5)),
            ),
            child: const Row(
              children: [
                Text('⚡', style: TextStyle(fontSize: 18)),
                SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Self-Checkout Ready', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: ZippyraColors.successGreen)),
                      Text('Exit gate auto-verification active', style: TextStyle(fontSize: 9, color: ZippyraColors.textSecondary)),
                    ],
                  ),
                ),
                Text('VALID', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, color: ZippyraColors.successGreen)),
              ],
            ),
          ),

          // Items List
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 14),
              itemCount: _cartItems.length,
              itemBuilder: (context, index) {
                final item = _cartItems[index];

                return Container(
                  margin: const EdgeInsets.only(bottom: 10),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(color: ZippyraColors.border),
                  ),
                  child: Row(
                    children: [
                      Text(item['emoji'] as String, style: const TextStyle(fontSize: 32)),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item['name'] as String,
                              style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: ZippyraColors.textPrimary),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              item['weight'] as String,
                              style: const TextStyle(fontSize: 10, color: ZippyraColors.textSecondary),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '₹${item['price']}',
                              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w900, color: ZippyraColors.textPrimary),
                            ),
                          ],
                        ),
                      ),

                      // Qty Controller
                      Container(
                        decoration: BoxDecoration(
                          border: Border.all(color: ZippyraColors.border),
                          borderRadius: BorderRadius.circular(9),
                        ),
                        child: Row(
                          children: [
                            GestureDetector(
                              onTap: () {
                                setState(() {
                                  if ((item['qty'] as int) > 1) {
                                    item['qty'] = (item['qty'] as int) - 1;
                                  } else {
                                    _cartItems.removeAt(index);
                                  }
                                });
                              },
                              child: Container(
                                width: 26,
                                height: 26,
                                decoration: const BoxDecoration(
                                  color: Color(0xFFE8F1FB),
                                  borderRadius: BorderRadius.horizontal(left: Radius.circular(8)),
                                ),
                                child: const Center(child: Text('−', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: ZippyraColors.primaryBlue))),
                              ),
                            ),
                            SizedBox(
                              width: 28,
                              child: Center(
                                child: Text('${item['qty']}', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
                              ),
                            ),
                            GestureDetector(
                              onTap: () {
                                setState(() => item['qty'] = (item['qty'] as int) + 1);
                              },
                              child: Container(
                                width: 26,
                                height: 26,
                                decoration: const BoxDecoration(
                                  color: Color(0xFFE8F1FB),
                                  borderRadius: BorderRadius.horizontal(right: Radius.circular(8)),
                                ),
                                child: const Center(child: Text('+', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: ZippyraColors.primaryBlue))),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),

          // Bill Summary & Checkout Bar
          Container(
            padding: const EdgeInsets.all(16),
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
              boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 10, offset: Offset(0, -4))],
            ),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Item Subtotal', style: TextStyle(fontSize: 12, color: ZippyraColors.textSecondary)),
                    Text('₹$_totalPrice', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: ZippyraColors.textPrimary)),
                  ],
                ),
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: const [
                    Text('Store Exit Verification', style: TextStyle(fontSize: 12, color: ZippyraColors.textSecondary)),
                    Text('FREE', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: ZippyraColors.successGreen)),
                  ],
                ),
                const Divider(height: 20, color: ZippyraColors.border),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('To Pay', style: TextStyle(fontSize: 10, color: ZippyraColors.textSecondary)),
                        Text('₹$_totalPrice', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w900, color: ZippyraColors.textPrimary)),
                      ],
                    ),
                    SizedBox(
                      height: 46,
                      width: 180,
                      child: ElevatedButton(
                        onPressed: _cartItems.isEmpty ? null : () => context.push('/checkout'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: ZippyraColors.primaryBlue,
                          foregroundColor: Colors.white,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          elevation: 2,
                        ),
                        child: const Text('Checkout →', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}


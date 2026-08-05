import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class CashPaymentScreen extends StatelessWidget {
  const CashPaymentScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Cashier Cash Override')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Cash Checkout (Non-App Customer)', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            const ZCard(
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [Text('Cart Total:'), Text('₹68.00', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18))],
                  ),
                  Divider(),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [Text('Cash Received:'), Text('₹100.00', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18))],
                  ),
                  Divider(),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [Text('Change Return:'), Text('₹32.00', style: TextStyle(color: ZippyraColors.successGreen, fontWeight: FontWeight.bold, fontSize: 18))],
                  ),
                ],
              ),
            ),
            const Spacer(),
            ZButton(
              label: 'Complete Cash Order & Issue Invoice',
              type: ZButtonType.green,
              onPressed: () {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Order #ZP-CASH-991 Created Successfully!')),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

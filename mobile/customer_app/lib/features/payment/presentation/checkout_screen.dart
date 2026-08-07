import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class CheckoutScreen extends StatelessWidget {
  const CheckoutScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Checkout')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Payment Method', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            ZCard(
              child: RadioListTile(
                value: 1,
                groupValue: 1,
                onChanged: (val) {},
                title: const Text('Razorpay UPI / Google Pay / PhonePe', style: TextStyle(fontWeight: FontWeight.bold)),
                secondary: const Icon(Icons.account_balance_wallet, color: ZippyraColors.primaryBlue),
              ),
            ),
            const Spacer(),
            ZButton(
              label: 'Pay ₹68 via Razorpay',
              type: ZButtonType.green,
              onPressed: () => context.push('/payment_processing'),
            ),
          ],
        ),
      ),
    );
  }
}

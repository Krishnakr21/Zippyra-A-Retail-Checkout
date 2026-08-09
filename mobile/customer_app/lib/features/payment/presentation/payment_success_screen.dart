import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';

class PaymentSuccessScreen extends StatelessWidget {
  const PaymentSuccessScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.check_circle_rounded, size: 100, color: ZippyraColors.successGreen),
            const SizedBox(height: 24),
            const Text('Payment Successful!', style: TextStyle(fontSize: 26, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            const Text('Order #ZP-9821 verified', style: TextStyle(color: ZippyraColors.textSecondary, fontSize: 16)),
            const SizedBox(height: 32),
            ZButton(
              label: 'Generate Exit QR Pass',
              type: ZButtonType.green,
              onPressed: () => context.go('/exit'),
            ),
          ],
        ),
      ),
    );
  }
}

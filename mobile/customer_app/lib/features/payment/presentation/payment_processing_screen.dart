import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class PaymentProcessingScreen extends StatefulWidget {
  const PaymentProcessingScreen({super.key});

  @override
  State<PaymentProcessingScreen> createState() => _PaymentProcessingScreenState();
}

class _PaymentProcessingScreenState extends State<PaymentProcessingScreen> {
  @override
  void initState() {
    super.initState();
    Future.delayed(const Duration(seconds: 3), () {
      if (mounted) context.go('/payment_success');
    });
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircularProgressIndicator(color: ZippyraColors.primaryBlue),
            SizedBox(height: 24),
            Text('Processing Payment...', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
            SizedBox(height: 8),
            Text('Confirming transaction with payment gateway', style: TextStyle(color: ZippyraColors.textSecondary)),
          ],
        ),
      ),
    );
  }
}

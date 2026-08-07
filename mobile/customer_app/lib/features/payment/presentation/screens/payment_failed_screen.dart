import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

class PaymentFailedScreen extends StatelessWidget {
  final String reason;

  const PaymentFailedScreen({
    super.key,
    required this.reason,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                width: 90,
                height: 90,
                decoration: BoxDecoration(
                  color: Colors.red[100],
                  shape: BoxShape.circle,
                ),
                child: Icon(Icons.close, size: 56, color: Colors.red[700]),
              ),
              const SizedBox(height: 24),
              const Text(
                'Payment Failed',
                style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              Text(
                reason.isNotEmpty ? reason : 'Your transaction could not be completed by the bank.',
                textAlign: TextAlign.center,
                style: const TextStyle(color: Colors.grey, fontSize: 14),
              ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: ZippyraColors.primary,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  ),
                  onPressed: () {
                    context.go('/cart');
                  },
                  child: const Text(
                    'Try Again',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

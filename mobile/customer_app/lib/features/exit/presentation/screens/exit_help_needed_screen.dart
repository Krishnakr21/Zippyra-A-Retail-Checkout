import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

class ExitHelpNeededScreen extends StatelessWidget {
  const ExitHelpNeededScreen({super.key});

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
                  color: Colors.blue[50],
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.help_outline, size: 56, color: Colors.blue),
              ),
              const SizedBox(height: 24),
              const Text(
                'Staff Assistance Needed',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              const Text(
                'Something went wrong at the exit scanner — a staff member will help you right away.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey, height: 1.4, fontSize: 14),
              ),
              const SizedBox(height: 32),
              Card(
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                child: const Padding(
                  padding: EdgeInsets.all(16.0),
                  child: Row(
                    children: [
                      Icon(Icons.info_outline, color: ZippyraColors.primary, size: 28),
                      SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Store security staff can verify your receipt and assist with opening the exit gate.',
                          style: TextStyle(fontSize: 13, color: Colors.black87),
                        ),
                      ),
                    ],
                  ),
                ),
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
                    context.go('/orders');
                  },
                  child: const Text(
                    'View My Orders',
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

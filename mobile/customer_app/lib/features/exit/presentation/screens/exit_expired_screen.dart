import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

class ExitExpiredScreen extends StatelessWidget {
  const ExitExpiredScreen({super.key});

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
                  color: Colors.orange[50],
                  shape: BoxShape.circle,
                ),
                child: Icon(Icons.timer_off_outlined, size: 56, color: Colors.orange[800]),
              ),
              const SizedBox(height: 24),
              const Text(
                'Exit Pass Expired',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              const Text(
                'Your exit code expired. Please see a staff member for assistance at the exit gate.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey, height: 1.4, fontSize: 14),
              ),
              const SizedBox(height: 32),
              Card(
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                color: Colors.amber[50],
                child: const Padding(
                  padding: EdgeInsets.all(16.0),
                  child: Row(
                    children: [
                      Icon(Icons.person_pin_circle_outlined, color: Colors.amber, size: 32),
                      SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Show your payment receipt in the Orders tab to store security staff to open the gate.',
                          style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.black87),
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

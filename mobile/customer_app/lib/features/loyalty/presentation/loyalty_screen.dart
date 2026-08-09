import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbadge.dart';

class LoyaltyScreen extends StatelessWidget {
  const LoyaltyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Zippyra Rewards')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                gradient: const LinearGradient(colors: [ZippyraColors.accentOrange, Colors.orangeAccent]),
                borderRadius: BorderRadius.circular(16),
              ),
              child: const Column(
                children: [
                  ZBadge(text: 'GOLD MEMBER', color: Colors.white),
                  SizedBox(height: 12),
                  Text('450 Points Available', style: TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold)),
                  SizedBox(height: 4),
                  Text('Redeem ₹45 on your next checkout', style: TextStyle(color: Colors.white70)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

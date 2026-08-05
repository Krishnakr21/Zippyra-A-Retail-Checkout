import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class PriceCheckScreen extends StatelessWidget {
  const PriceCheckScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Price Lookup')),
      body: const Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          children: [
            ZCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Amul Taaza Milk 1L', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                  SizedBox(height: 4),
                  Text('Price: ₹68.00 (MSRP ₹70)', style: TextStyle(color: ZippyraColors.primaryBlue, fontSize: 20, fontWeight: FontWeight.bold)),
                  SizedBox(height: 8),
                  Text('Active Offer: 20% off on fresh dairy items', style: TextStyle(color: ZippyraColors.successGreen, fontWeight: FontWeight.w600)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

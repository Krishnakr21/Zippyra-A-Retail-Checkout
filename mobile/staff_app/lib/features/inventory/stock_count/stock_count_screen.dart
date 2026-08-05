import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zcard.dart';
import 'package:zippyra_core/widgets/zbadge.dart';

class StockCountScreen extends StatefulWidget {
  const StockCountScreen({super.key});

  @override
  State<StockCountScreen> createState() => _StockCountScreenState();
}

class _StockCountScreenState extends State<StockCountScreen> {
  final TextEditingController _qtyController = TextEditingController(text: '40');

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Offline Stock Count')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            ZCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text('Amul Taaza Milk 1L', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                      ZBadge(text: 'VARIANCE -2', color: ZippyraColors.errorRed),
                    ],
                  ),
                  const SizedBox(height: 8),
                  const Text('Barcode: 890123456789', style: TextStyle(color: ZippyraColors.textSecondary)),
                  const Text('System Expected: 42 Units', style: TextStyle(color: ZippyraColors.primaryBlue, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _qtyController,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      labelText: 'Counted Physical Quantity',
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                  ),
                ],
              ),
            ),
            const Spacer(),
            ZButton(
              label: 'Submit Count (Enqueue Offline)',
              type: ZButtonType.primary,
              onPressed: () {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Stock count queued for sync!')),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

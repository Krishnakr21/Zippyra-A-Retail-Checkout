import 'package:flutter/material.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class GrnScreen extends StatelessWidget {
  const GrnScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Goods Received Note (GRN)')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Select Purchase Order (PO)', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            const ZCard(
              child: ListTile(
                title: Text('PO #PO-2026-8891', style: TextStyle(fontWeight: FontWeight.bold)),
                subtitle: Text('Supplier: Amul Dairy Corp • 50 Units'),
                trailing: Icon(Icons.arrow_forward_ios, size: 16),
              ),
            ),
            const Spacer(),
            ZButton(
              label: 'Scan Incoming PO Items',
              type: ZButtonType.green,
              onPressed: () {},
            ),
          ],
        ),
      ),
    );
  }
}

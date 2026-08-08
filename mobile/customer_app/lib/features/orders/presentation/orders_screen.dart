import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class OrdersScreen extends StatelessWidget {
  const OrdersScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Order History')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: const [
          ZCard(
            child: ListTile(
              title: Text('Order #ZP-9821', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: Text('28 July 2026 • ₹68 (1 Item)'),
              trailing: Icon(Icons.chevron_right, color: ZippyraColors.primaryBlue),
            ),
          ),
        ],
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class CustomerAssistScreen extends StatelessWidget {
  const CustomerAssistScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Customer Assist')),
      body: const Padding(
        padding: EdgeInsets.all(16.0),
        child: Column(
          children: [
            ZCard(
              child: ListTile(
                leading: Icon(Icons.person_search, color: ZippyraColors.primaryBlue, size: 36),
                title: Text('Customer: Rahul Verma', style: TextStyle(fontWeight: FontWeight.bold)),
                subtitle: Text('Active Session • Cart: 2 Items (₹136)'),
                trailing: Icon(Icons.visibility, color: ZippyraColors.primaryBlue),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

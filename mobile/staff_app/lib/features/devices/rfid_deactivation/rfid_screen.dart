import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';

class RfidDeactivationScreen extends StatelessWidget {
  const RfidDeactivationScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('RFID Counter Deactivation')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.nfc, size: 100, color: ZippyraColors.accentOrange),
            const SizedBox(height: 24),
            const Text('Hold Tag Near Counter Reader', style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            const Text('Deactivates anti-theft RFID tag on paid items', style: TextStyle(color: ZippyraColors.textSecondary)),
            const SizedBox(height: 32),
            ZButton(
              label: 'Deactivate Tag (EPC #99812)',
              type: ZButtonType.orange,
              onPressed: () {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('RFID Tag Deactivated Successfully!')),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';

class UpdateRequiredScreen extends StatelessWidget {
  const UpdateRequiredScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.system_update_rounded, size: 90, color: ZippyraColors.accentOrange),
            const SizedBox(height: 24),
            const Text(
              'App Update Required',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            const Text(
              'A critical update is required to continue shopping securely at Zippyra stores.',
              textAlign: TextAlign.center,
              style: TextStyle(color: ZippyraColors.textSecondary, fontSize: 15),
            ),
            const SizedBox(height: 32),
            ZButton(
              label: 'Update Now',
              type: ZButtonType.orange,
              onPressed: () {},
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';

class ExitScreen extends StatelessWidget {
  const ExitScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Exit Turnstile Pass')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          children: [
            const Text('Scan at Exit Turnstile Gate', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            const Text('Valid for 10 minutes (Ed25519 single-use token)', style: TextStyle(color: ZippyraColors.textSecondary)),
            const SizedBox(height: 32),
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 15)],
              ),
              child: const Icon(Icons.qr_code_2_rounded, size: 220, color: ZippyraColors.primaryBlue),
            ),
            const Spacer(),
            ZButton(
              label: 'Done & Back to Home',
              type: ZButtonType.ghost,
              onPressed: () => context.go('/home'),
            ),
          ],
        ),
      ),
    );
  }
}

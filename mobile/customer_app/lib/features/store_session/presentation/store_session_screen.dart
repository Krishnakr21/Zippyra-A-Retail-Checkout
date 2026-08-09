import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class StoreSessionScreen extends StatelessWidget {
  const StoreSessionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Select Store')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          children: [
            const Icon(Icons.storefront_outlined, size: 80, color: ZippyraColors.primaryBlue),
            const SizedBox(height: 16),
            const Text(
              'Nearby Store Detected',
              style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            const Text(
              'Bind your session to start scanning items.',
              style: TextStyle(color: ZippyraColors.textSecondary),
            ),
            const SizedBox(height: 24),
            ZCard(
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: ZippyraColors.primaryBlue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.store, color: ZippyraColors.primaryBlue, size: 32),
                  ),
                  const SizedBox(width: 16),
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Zippyra Superstore - Indiranagar', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        SizedBox(height: 4),
                        Text('100m away • Open until 11:00 PM', style: TextStyle(color: ZippyraColors.textSecondary, fontSize: 13)),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const Spacer(),
            ZButton(
              label: 'Enter Store & Sync Catalog',
              onPressed: () => context.go('/home'),
            ),
          ],
        ),
      ),
    );
  }
}

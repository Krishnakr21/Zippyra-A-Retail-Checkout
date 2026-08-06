import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

class StoreClosedScreen extends StatelessWidget {
  final String? message;

  const StoreClosedScreen({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Store Closed')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.access_time_filled, size: 72, color: Colors.orange),
              const SizedBox(height: 24),
              const Text(
                'Store is Currently Closed',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                message ?? 'This store is outside operating hours or undergoing maintenance.',
                textAlign: TextAlign.center,
                style: const TextStyle(color: ZippyraColors.textSecondary),
              ),
              const SizedBox(height: 32),
              ZButton(
                label: 'View Other Nearby Stores',
                onPressed: () => context.go('/store/list'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

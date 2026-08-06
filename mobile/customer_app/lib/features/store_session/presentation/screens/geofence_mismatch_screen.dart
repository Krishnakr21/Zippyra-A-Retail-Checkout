import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';

class GeofenceMismatchScreen extends StatelessWidget {
  final String? message;

  const GeofenceMismatchScreen({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Geofence Error')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.location_off, size: 72, color: Colors.red),
              const SizedBox(height: 24),
              const Text(
                'Outside Store Geofence',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                message ?? 'Move closer to the store entrance and try again.',
                textAlign: TextAlign.center,
                style: const TextStyle(color: ZippyraColors.textSecondary),
              ),
              const SizedBox(height: 32),
              ZButton(
                label: 'Re-Scan Entrance QR',
                onPressed: () => context.go('/store/scan'),
              ),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: () => context.go('/store/list'),
                child: const Text('Back to Store List'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

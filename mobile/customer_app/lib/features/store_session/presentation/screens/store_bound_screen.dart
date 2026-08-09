import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/store_session_bloc.dart';

class StoreBoundScreen extends StatefulWidget {
  const StoreBoundScreen({super.key});

  @override
  State<StoreBoundScreen> createState() => _StoreBoundScreenState();
}

class _StoreBoundScreenState extends State<StoreBoundScreen> {
  @override
  void initState() {
    super.initState();
    // Auto navigate to home/scan after brief success display
    Future.delayed(const Duration(seconds: 2), () {
      if (mounted) {
        context.go('/home');
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocBuilder<StoreSessionBloc, StoreSessionState>(
        builder: (context, state) {
          final storeName = state is StoreSessionActive ? state.session.storeName : 'Zippyra Store';
          final catalogVer = state is StoreSessionActive ? state.session.catalogVersion : 1;

          return SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const CircleAvatar(
                    radius: 40,
                    backgroundColor: Colors.green,
                    child: Icon(Icons.check, size: 48, color: Colors.white),
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'Welcome to $storeName!',
                    style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'You are successfully checked in. Start scanning items or browsing catalog.',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: ZippyraColors.textSecondary),
                  ),
                  const SizedBox(height: 24),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                      const SizedBox(width: 8),
                      Text('Syncing store catalog (v$catalogVer)...', style: const TextStyle(color: Colors.grey, fontSize: 12)),
                    ],
                  ),
                  const SizedBox(height: 32),
                  ZButton(
                    label: 'Start Self-Checkout',
                    onPressed: () => context.go('/home'),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

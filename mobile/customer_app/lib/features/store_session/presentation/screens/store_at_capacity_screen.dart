import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/store_session_bloc.dart';

class StoreAtCapacityScreen extends StatefulWidget {
  final String? message;

  const StoreAtCapacityScreen({super.key, this.message});

  @override
  State<StoreAtCapacityScreen> createState() => _StoreAtCapacityScreenState();
}

class _StoreAtCapacityScreenState extends State<StoreAtCapacityScreen> {
  int _retrySeconds = 60;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _startTimer();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _startTimer() {
    setState(() => _retrySeconds = 60);
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_retrySeconds > 0) {
        setState(() => _retrySeconds--);
      } else {
        timer.cancel();
        context.read<StoreSessionBloc>().add(CapacityRetryTimerFired());
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Store Full')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.people_alt, size: 72, color: Colors.amber),
              const SizedBox(height: 24),
              const Text(
                'Store Reached Maximum Capacity',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                widget.message ?? 'To ensure safety and comfort, entry is temporarily paused.',
                textAlign: TextAlign.center,
                style: const TextStyle(color: ZippyraColors.textSecondary),
              ),
              const SizedBox(height: 24),
              Text(
                _retrySeconds > 0 ? 'Auto Retrying in ${_retrySeconds}s...' : 'Retrying entry...',
                style: const TextStyle(fontWeight: FontWeight.bold, color: ZippyraColors.primaryBlue),
              ),
              const SizedBox(height: 32),
              ZButton(
                label: 'Retry Entry Now',
                onPressed: () {
                  context.read<StoreSessionBloc>().add(CapacityRetryTimerFired());
                  context.go('/store/binding');
                },
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

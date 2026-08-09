import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../../injection_container.dart';
import '../../../feedback/data/feedback_service.dart';
import '../../../feedback/presentation/feedback_modal.dart';
import '../../../store_session/presentation/bloc/store_session_bloc.dart';

class ExitSuccessScreen extends StatefulWidget {
  const ExitSuccessScreen({super.key});

  @override
  State<ExitSuccessScreen> createState() => _ExitSuccessScreenState();
}

class _ExitSuccessScreenState extends State<ExitSuccessScreen> {
  Timer? _navTimer;

  @override
  void initState() {
    super.initState();
    // Immediate client-side session cleanup
    context.read<StoreSessionBloc>().add(const UnbindRequested(reason: 'customer_exit'));

    // Check feedback gating (1 per 3 completed orders)
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      try {
        final feedbackService = sl<FeedbackService>();
        final shouldShow = await feedbackService.incrementOrderAndCheckGating();
        if (shouldShow && mounted) {
          _navTimer?.cancel();
          FeedbackModal.show(context, feedbackService: feedbackService);
        }
      } catch (_) {}
    });

    // Auto-navigate to home after 3s if modal not shown
    _navTimer = Timer(const Duration(seconds: 3), () {
      if (mounted) {
        context.go('/');
      }
    });
  }

  @override
  void dispose() {
    _navTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                width: 100,
                height: 100,
                decoration: const BoxDecoration(
                  color: Colors.green,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.sensor_door_outlined, size: 64, color: Colors.white),
              ),
              const SizedBox(height: 24),
              const Text(
                'Gate Opened!',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              const Text(
                'Thank you for shopping with us! Have a great day.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey, fontSize: 16),
              ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: ZippyraColors.primary,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  ),
                  onPressed: () {
                    _navTimer?.cancel();
                    context.go('/');
                  },
                  child: const Text(
                    'Done',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

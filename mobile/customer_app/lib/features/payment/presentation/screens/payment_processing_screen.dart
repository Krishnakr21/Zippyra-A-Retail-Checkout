import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/payment_bloc.dart';

class PaymentProcessingScreen extends StatefulWidget {
  final String paymentId;

  const PaymentProcessingScreen({
    super.key,
    required this.paymentId,
  });

  @override
  State<PaymentProcessingScreen> createState() => _PaymentProcessingScreenState();
}

class _PaymentProcessingScreenState extends State<PaymentProcessingScreen> {
  Timer? _pollTimer;

  @override
  void initState() {
    super.initState();
    _startPolling();
  }

  void _startPolling() {
    _pollTimer = Timer.periodic(const Duration(seconds: 3), (_) {
      if (mounted) {
        context.read<PaymentBloc>().add(PaymentStatusPollTicked(widget.paymentId));
      }
    });
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false, // Non-dismissible while polling is active
      child: Scaffold(
        body: BlocConsumer<PaymentBloc, PaymentState>(
          listener: (context, state) {
            if (state is PaymentSuccess) {
              _pollTimer?.cancel();
              context.go('/payment/success', extra: {'payment_id': state.paymentId});
            } else if (state is PaymentFailed) {
              _pollTimer?.cancel();
              context.go('/payment/failed', extra: {'reason': state.reason});
            } else if (state is PaymentPendingTimeout) {
              _pollTimer?.cancel();
              context.go('/payment/pending-timeout');
            }
          },
          builder: (context, state) {
            int attempts = 30;
            if (state is PaymentProcessing) {
              attempts = state.attemptsRemaining;
            }

            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const SizedBox(
                      width: 70,
                      height: 70,
                      child: CircularProgressIndicator(
                        color: ZippyraColors.primary,
                        strokeWidth: 4,
                      ),
                    ),
                    const SizedBox(height: 32),
                    const Text(
                      'Processing Your Payment...',
                      style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Please do not press back or close the application. Verifying with bank & UPI gateway...',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey, height: 1.4),
                    ),
                    const SizedBox(height: 24),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                      decoration: BoxDecoration(
                        color: Colors.blue[50],
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        'Checking status... ($attempts attempts remaining)',
                        style: TextStyle(color: Colors.blue[800], fontSize: 13, fontWeight: FontWeight.bold),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}

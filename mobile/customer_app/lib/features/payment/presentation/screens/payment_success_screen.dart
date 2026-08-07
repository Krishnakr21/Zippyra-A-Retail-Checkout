import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../orders/presentation/cubit/order_exit_cubit.dart';

class PaymentSuccessScreen extends StatelessWidget {
  final String paymentId;
  final String storeId;

  const PaymentSuccessScreen({
    super.key,
    required this.paymentId,
    this.storeId = 'store-1',
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: BlocConsumer<OrderExitCubit, OrderExitState>(
            listener: (context, state) {
              if (state is OrderExitLoaded) {
                context.go('/exit/qr', extra: {
                  'payment_id': paymentId,
                  'exit_token': state.exitToken.token,
                  'expires_at': state.exitToken.expiresAt,
                });
              }
            },
            builder: (context, state) {
              return Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Container(
                    width: 90,
                    height: 90,
                    decoration: const BoxDecoration(
                      color: Colors.green,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.check, size: 56, color: Colors.white),
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    'Payment Successful!',
                    style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Payment Reference: ${paymentId.substring(0, paymentId.length.clamp(0, 18))}',
                    style: const TextStyle(color: Colors.grey, fontSize: 13),
                  ),
                  const SizedBox(height: 24),
                  Card(
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: const Padding(
                      padding: EdgeInsets.all(16.0),
                      child: Row(
                        children: [
                          Icon(Icons.qr_code, color: ZippyraColors.primary, size: 28),
                          SizedBox(width: 12),
                          Expanded(
                            child: Text(
                              'Your Exit Pass QR is generated. Show it at the store exit gate.',
                              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  if (state is OrderExitError) ...[
                    const SizedBox(height: 12),
                    Text(
                      'Exit pass token pending: ${state.message}',
                      style: const TextStyle(color: Colors.red, fontSize: 12),
                    ),
                  ],
                  const Spacer(),
                  SizedBox(
                    width: double.infinity,
                    height: 50,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: ZippyraColors.primary,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      onPressed: state is OrderExitLoading
                          ? null
                          : () {
                              context.read<OrderExitCubit>().fetchExitToken(storeId);
                            },
                      child: state is OrderExitLoading
                          ? const CircularProgressIndicator(color: Colors.white)
                          : const Text(
                              'View Exit Pass QR',
                              style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                            ),
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}

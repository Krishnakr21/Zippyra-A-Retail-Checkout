import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/payment/domain/entities/payment_intent.dart';
import 'package:customer_app/features/payment/domain/entities/payment_status_entity.dart';
import 'package:customer_app/features/payment/domain/entities/split_estimate.dart';
import 'package:customer_app/features/payment/domain/repositories/payment_repository.dart';
import 'package:customer_app/features/payment/domain/usecases/check_pending_payment_use_case.dart';
import 'package:customer_app/features/payment/domain/usecases/estimate_split_use_case.dart';
import 'package:customer_app/features/payment/domain/usecases/get_payment_status_use_case.dart';
import 'package:customer_app/features/payment/domain/usecases/initiate_payment_use_case.dart';
import 'package:customer_app/features/payment/presentation/bloc/payment_bloc.dart';
import 'package:customer_app/features/payment/presentation/screens/payment_processing_screen.dart';

class MockPaymentRepo implements PaymentRepository {
  @override
  Future<SplitEstimate> estimateSplit({required String checkoutSessionId, required int pointsToRedeem}) async {
    return const SplitEstimate(originalTotalPaise: 50000, loyaltyDiscountPaise: 0, payableAmountPaise: 50000, pointsBalance: 500);
  }

  @override
  Future<PaymentIntent> initiatePayment({required String checkoutSessionId, required String method, int pointsToRedeem = 0, String? playIntegrityToken}) async {
    return PaymentIntent(paymentId: 'pay-1', gateway: 'rzp', gatewayOrderId: 'o1', gatewayKeyId: 'k1', payableAmountPaise: 50000, expiresAt: DateTime.now());
  }

  @override
  Future<PaymentStatusEntity> getPaymentStatus(String paymentId) async {
    return const PaymentStatusEntity(status: 'PENDING');
  }

  @override
  Future<void> savePendingPaymentId(String paymentId) async {}

  @override
  Future<String?> getPendingPaymentId() async => null;

  @override
  Future<void> clearPendingPaymentId() async {}
}

void main() {
  testWidgets('PaymentProcessingScreen blocks system back button via PopScope(canPop: false)', (tester) async {
    final repo = MockPaymentRepo();
    final bloc = PaymentBloc(
      estimateSplitUseCase: EstimateSplitUseCase(repo),
      initiatePaymentUseCase: InitiatePaymentUseCase(repo),
      getPaymentStatusUseCase: GetPaymentStatusUseCase(repo),
      checkPendingPaymentUseCase: CheckPendingPaymentUseCase(repo),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<PaymentBloc>.value(
          value: bloc,
          child: const PaymentProcessingScreen(paymentId: 'pay-proc-123'),
        ),
      ),
    );

    final popScopeFinder = find.byType(PopScope);
    expect(popScopeFinder, findsOneWidget);

    final popScopeWidget = tester.widget<PopScope>(popScopeFinder);
    expect(popScopeWidget.canPop, isFalse);

    bloc.close();
  });
}

import 'package:bloc_test/bloc_test.dart';
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

class FakePaymentRepository implements PaymentRepository {
  String? savedPendingId;
  int pollCount = 0;
  String targetPollStatus = 'PENDING';

  @override
  Future<SplitEstimate> estimateSplit({required String checkoutSessionId, required int pointsToRedeem}) async {
    return SplitEstimate(
      originalTotalPaise: 50000,
      loyaltyDiscountPaise: pointsToRedeem * 1,
      payableAmountPaise: 50000 - (pointsToRedeem * 1),
      pointsBalance: 500,
    );
  }

  @override
  Future<PaymentIntent> initiatePayment({required String checkoutSessionId, required String method, int pointsToRedeem = 0, String? playIntegrityToken}) async {
    return PaymentIntent(
      paymentId: 'pay-test-100',
      gateway: 'razorpay',
      gatewayOrderId: 'order_100',
      gatewayKeyId: 'key_100',
      payableAmountPaise: 50000,
      expiresAt: DateTime.now().add(const Duration(minutes: 10)),
    );
  }

  @override
  Future<PaymentStatusEntity> getPaymentStatus(String paymentId) async {
    pollCount++;
    if (targetPollStatus == 'SEQUENCE') {
      if (pollCount >= 3) {
        return const PaymentStatusEntity(status: 'CAPTURED');
      }
      return const PaymentStatusEntity(status: 'PENDING');
    }
    return PaymentStatusEntity(status: targetPollStatus, failureReason: targetPollStatus == 'FAILED' ? 'Bank Declined' : null);
  }

  @override
  Future<void> savePendingPaymentId(String paymentId) async {
    savedPendingId = paymentId;
  }

  @override
  Future<String?> getPendingPaymentId() async {
    return savedPendingId;
  }

  @override
  Future<void> clearPendingPaymentId() async {
    savedPendingId = null;
  }
}

void main() {
  late FakePaymentRepository repo;
  late PaymentBloc bloc;

  setUp(() {
    repo = FakePaymentRepository();
    bloc = PaymentBloc(
      estimateSplitUseCase: EstimateSplitUseCase(repo),
      initiatePaymentUseCase: InitiatePaymentUseCase(repo),
      getPaymentStatusUseCase: GetPaymentStatusUseCase(repo),
      checkPendingPaymentUseCase: CheckPendingPaymentUseCase(repo),
    );
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is PaymentInitial', () {
    expect(bloc.state, isA<PaymentInitial>());
  });

  blocTest<PaymentBloc, PaymentState>(
    'PaymentInitiateRequested success immediately transitions to PaymentProcessing AND triggers SecureStorage write',
    build: () => bloc,
    act: (bloc) => bloc.add(const PaymentInitiateRequested('sess-100')),
    expect: () => [
      isA<PaymentInitiating>(),
      isA<PaymentProcessing>().having((s) => s.paymentId, 'paymentId', 'pay-test-100'),
    ],
    verify: (_) {
      // SAFETY-CRITICAL assertion: pending_payment_id MUST be saved immediately
      expect(repo.savedPendingId, equals('pay-test-100'));
    },
  );

  blocTest<PaymentBloc, PaymentState>(
    'Poll sequence pending -> pending -> captured transitions correctly to PaymentSuccess',
    build: () {
      repo.targetPollStatus = 'SEQUENCE';
      return bloc;
    },
    act: (bloc) async {
      bloc.add(const PaymentStatusPollTicked('pay-test-100')); // count 1 -> PENDING
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(const PaymentStatusPollTicked('pay-test-100')); // count 2 -> PENDING
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(const PaymentStatusPollTicked('pay-test-100')); // count 3 -> CAPTURED
    },
    expect: () => [
      isA<PaymentProcessing>().having((s) => s.attemptsRemaining, 'attemptsRemaining', 29),
      isA<PaymentProcessing>().having((s) => s.attemptsRemaining, 'attemptsRemaining', 28),
      isA<PaymentSuccess>().having((s) => s.paymentId, 'paymentId', 'pay-test-100'),
    ],
  );

  blocTest<PaymentBloc, PaymentState>(
    'PendingPaymentCheckRequested with stored ID resolving to failed clears key and emits PaymentFailed',
    build: () {
      repo.savedPendingId = 'pay-failed-999';
      repo.targetPollStatus = 'FAILED';
      return bloc;
    },
    act: (bloc) => bloc.add(PendingPaymentCheckRequested()),
    expect: () => [
      isA<PaymentRecoveryChecking>(),
      isA<PaymentFailed>().having((s) => s.reason, 'reason', 'Bank Declined'),
    ],
    verify: (_) {
      expect(repo.savedPendingId, isNull);
    },
  );
}

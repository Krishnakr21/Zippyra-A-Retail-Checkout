import 'dart:async';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:stream_transform/stream_transform.dart';
import 'package:zippyra_core/zippyra_core.dart';

import '../../data/razorpay_service.dart';
import '../../domain/entities/payment_intent.dart';
import '../../domain/usecases/check_pending_payment_use_case.dart';
import '../../domain/usecases/estimate_split_use_case.dart';
import '../../domain/usecases/get_payment_status_use_case.dart';
import '../../domain/usecases/initiate_payment_use_case.dart';

import '../../../../core/services/play_integrity_service.dart';

part 'payment_event.dart';
part 'payment_state.dart';

EventTransformer<E> debounce<E>(Duration duration) {
  return (events, mapper) => events.debounce(duration).switchMap(mapper);
}

class PaymentBloc extends Bloc<PaymentEvent, PaymentState> {
  final EstimateSplitUseCase estimateSplitUseCase;
  final InitiatePaymentUseCase initiatePaymentUseCase;
  final GetPaymentStatusUseCase getPaymentStatusUseCase;
  final CheckPendingPaymentUseCase checkPendingPaymentUseCase;
  final PlayIntegrityService? playIntegrityService;
  final RazorpayService? razorpayService;

  StreamSubscription? _razorpaySuccessSub;
  StreamSubscription? _razorpayErrorSub;

  PaymentBloc({
    required this.estimateSplitUseCase,
    required this.initiatePaymentUseCase,
    required this.getPaymentStatusUseCase,
    required this.checkPendingPaymentUseCase,
    this.playIntegrityService,
    this.razorpayService,
  }) : super(PaymentInitial()) {
    on<LoyaltyPointsSliderChanged>(
      _onLoyaltyPointsSliderChanged,
      transformer: debounce(const Duration(milliseconds: 300)),
    );
    on<PaymentMethodSelected>(_onPaymentMethodSelected);
    on<PaymentInitiateRequested>(_onPaymentInitiateRequested);
    on<PaymentStatusPollTicked>(_onPaymentStatusPollTicked);
    on<PaymentDeepLinkReceived>(_onPaymentDeepLinkReceived);
    on<PendingPaymentCheckRequested>(_onPendingPaymentCheckRequested);

    _listenToRazorpayEvents();
  }

  void _listenToRazorpayEvents() {
    if (razorpayService == null) return;
    _razorpaySuccessSub = razorpayService!.onSuccess.listen((response) {
      if (state is PaymentProcessing) {
        final paymentId = (state as PaymentProcessing).paymentId;
        add(PaymentStatusPollTicked(paymentId));
      }
    });
    _razorpayErrorSub = razorpayService!.onError.listen((response) {
      if (state is PaymentProcessing) {
        emit(PaymentFailed(response.message ?? 'Payment cancelled or failed at gateway'));
      }
    });
  }

  Future<void> _onLoyaltyPointsSliderChanged(
    LoyaltyPointsSliderChanged event,
    Emitter<PaymentState> emit,
  ) async {
    if (state is PaymentCheckoutReady) {
      final current = state as PaymentCheckoutReady;
      try {
        final estimate = await estimateSplitUseCase(
          checkoutSessionId: event.checkoutSessionId,
          pointsToRedeem: event.points,
        );
        emit(current.copyWith(
          pointsToRedeem: event.points,
          estimatedPayablePaise: estimate.payableAmountPaise,
          pointsBalance: estimate.pointsBalance,
        ));
      } catch (_) {}
    }
  }

  void _onPaymentMethodSelected(
    PaymentMethodSelected event,
    Emitter<PaymentState> emit,
  ) {
    if (state is PaymentCheckoutReady) {
      final current = state as PaymentCheckoutReady;
      emit(current.copyWith(selectedMethod: event.method));
    }
  }

  Future<void> _onPaymentInitiateRequested(
    PaymentInitiateRequested event,
    Emitter<PaymentState> emit,
  ) async {
    String method = 'UPI';
    int pointsToRedeem = 0;
    if (state is PaymentCheckoutReady) {
      final ready = state as PaymentCheckoutReady;
      method = ready.selectedMethod;
      pointsToRedeem = ready.pointsToRedeem;
    }

    emit(PaymentInitiating());
    try {
      final integrityToken = await playIntegrityService?.requestIntegrityToken(event.checkoutSessionId);

      // SAFETY-CRITICAL: Initiate call immediately writes pending_payment_id to SecureStorage
      final intent = await initiatePaymentUseCase(
        checkoutSessionId: event.checkoutSessionId,
        method: method,
        pointsToRedeem: pointsToRedeem,
        playIntegrityToken: integrityToken,
      );

      // Transition to PaymentProcessing FIRST
      emit(PaymentProcessing(
        paymentId: intent.paymentId,
        attemptsRemaining: 30,
      ));

      // Invoke Razorpay SDK checkout sheet if razorpayService is present
      if (razorpayService != null && intent.gatewayKeyId.isNotEmpty) {
        razorpayService!.open({
          'key': intent.gatewayKeyId,
          'amount': intent.payableAmountPaise,
          'name': 'Zippyra Store Checkout',
          'order_id': intent.gatewayOrderId,
          'prefill': {
            'contact': '',
            'email': '',
          },
          'external': {
            'wallets': ['paytm']
          }
        });
      }
    } catch (e) {
      emit(PaymentFailed(e.toString()));
    }
  }

  Future<void> _onPaymentStatusPollTicked(
    PaymentStatusPollTicked event,
    Emitter<PaymentState> emit,
  ) async {
    int remaining = 30;
    if (state is PaymentProcessing) {
      remaining = (state as PaymentProcessing).attemptsRemaining;
    }

    try {
      final status = await getPaymentStatusUseCase(event.paymentId);
      if (status.isCaptured) {
        emit(PaymentSuccess(event.paymentId));
      } else if (status.isFailed) {
        emit(PaymentFailed(status.failureReason ?? 'Payment failed'));
      } else {
        if (remaining > 1) {
          emit(PaymentProcessing(
            paymentId: event.paymentId,
            attemptsRemaining: remaining - 1,
          ));
        } else {
          emit(PaymentPendingTimeout());
        }
      }
    } catch (e) {
      if (remaining > 1) {
        emit(PaymentProcessing(
          paymentId: event.paymentId,
          attemptsRemaining: remaining - 1,
        ));
      } else {
        emit(PaymentPendingTimeout());
      }
    }
  }

  Future<void> _onPaymentDeepLinkReceived(
    PaymentDeepLinkReceived event,
    Emitter<PaymentState> emit,
  ) async {
    add(PaymentStatusPollTicked(event.paymentId));
  }

  Future<void> _onPendingPaymentCheckRequested(
    PendingPaymentCheckRequested event,
    Emitter<PaymentState> emit,
  ) async {
    emit(PaymentRecoveryChecking());
    final res = await checkPendingPaymentUseCase();
    if (res == null) {
      emit(PaymentInitial());
      return;
    }

    final paymentId = res['payment_id'] as String;
    final status = res['status'] as dynamic;

    if (status != null && status.isCaptured) {
      emit(PaymentSuccess(paymentId));
    } else if (status != null && status.isFailed) {
      emit(PaymentFailed(status.failureReason ?? 'Payment failed'));
    } else {
      emit(PaymentProcessing(paymentId: paymentId, attemptsRemaining: 30));
    }
  }

  @override
  Future<void> close() {
    _razorpaySuccessSub?.cancel();
    _razorpayErrorSub?.cancel();
    return super.close();
  }
}

import '../entities/payment_intent.dart';
import '../repositories/payment_repository.dart';

class InitiatePaymentUseCase {
  final PaymentRepository repository;

  InitiatePaymentUseCase(this.repository);

  Future<PaymentIntent> call({
    required String checkoutSessionId,
    required String method,
    int pointsToRedeem = 0,
    String? playIntegrityToken,
  }) async {
    final intent = await repository.initiatePayment(
      checkoutSessionId: checkoutSessionId,
      method: method,
      pointsToRedeem: pointsToRedeem,
      playIntegrityToken: playIntegrityToken,
    );
    // SAFETY-CRITICAL: Immediately persist pending_payment_id BEFORE opening Razorpay SDK / UPI app
    await repository.savePendingPaymentId(intent.paymentId);
    return intent;
  }
}

import '../entities/payment_intent.dart';
import '../entities/payment_status_entity.dart';
import '../entities/split_estimate.dart';

abstract class PaymentRepository {
  Future<SplitEstimate> estimateSplit({
    required String checkoutSessionId,
    required int pointsToRedeem,
  });

  Future<PaymentIntent> initiatePayment({
    required String checkoutSessionId,
    required String method,
    int pointsToRedeem = 0,
    String? playIntegrityToken,
  });

  Future<PaymentStatusEntity> getPaymentStatus(String paymentId);

  Future<void> savePendingPaymentId(String paymentId);
  Future<String?> getPendingPaymentId();
  Future<void> clearPendingPaymentId();
}

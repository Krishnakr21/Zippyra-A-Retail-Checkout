import '../entities/payment_status_entity.dart';
import '../repositories/payment_repository.dart';

class GetPaymentStatusUseCase {
  final PaymentRepository repository;

  GetPaymentStatusUseCase(this.repository);

  Future<PaymentStatusEntity> call(String paymentId) async {
    final status = await repository.getPaymentStatus(paymentId);
    if (status.isCaptured || status.isFailed) {
      await repository.clearPendingPaymentId();
    }
    return status;
  }
}

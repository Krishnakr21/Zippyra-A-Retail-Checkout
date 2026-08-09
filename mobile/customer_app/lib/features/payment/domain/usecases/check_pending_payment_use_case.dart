import '../entities/payment_status_entity.dart';
import '../repositories/payment_repository.dart';

class CheckPendingPaymentUseCase {
  final PaymentRepository repository;

  CheckPendingPaymentUseCase(this.repository);

  Future<Map<String, dynamic>?> call() async {
    final pendingId = await repository.getPendingPaymentId();
    if (pendingId == null || pendingId.isEmpty) {
      return null;
    }

    try {
      final status = await repository.getPaymentStatus(pendingId);
      if (status.isCaptured || status.isFailed) {
        await repository.clearPendingPaymentId();
      }
      return {
        'payment_id': pendingId,
        'status': status,
      };
    } catch (_) {
      return {
        'payment_id': pendingId,
        'status': const PaymentStatusEntity(status: 'PENDING'),
      };
    }
  }
}

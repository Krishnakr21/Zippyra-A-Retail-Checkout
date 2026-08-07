import '../entities/split_estimate.dart';
import '../repositories/payment_repository.dart';

class EstimateSplitUseCase {
  final PaymentRepository repository;

  EstimateSplitUseCase(this.repository);

  Future<SplitEstimate> call({
    required String checkoutSessionId,
    required int pointsToRedeem,
  }) {
    return repository.estimateSplit(
      checkoutSessionId: checkoutSessionId,
      pointsToRedeem: pointsToRedeem,
    );
  }
}

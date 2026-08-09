import 'package:zippyra_core/zippyra_core.dart';
import '../entities/exit_token.dart';
import '../repositories/orders_repository.dart';

class GetExitTokenUseCase {
  final OrdersRepository repository;
  final int maxRetries;
  final Duration retryDelay;

  GetExitTokenUseCase(
    this.repository, {
    this.maxRetries = 3,
    this.retryDelay = const Duration(seconds: 1),
  });

  Future<ExitToken> call({required String storeId}) async {
    Object? lastErr;
    for (int attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await repository.getExitToken(storeId: storeId);
      } catch (e) {
        lastErr = e;
        if (attempt < maxRetries) {
          await Future.delayed(retryDelay);
        }
      }
    }
    if (lastErr != null) {
      throw lastErr;
    }
    throw const ServerFailure('No pending exit pass found');
  }
}

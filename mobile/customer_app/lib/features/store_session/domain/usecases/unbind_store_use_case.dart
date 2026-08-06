import '../repositories/store_session_repository.dart';

class UnbindStoreUseCase {
  final StoreSessionRepository repository;

  const UnbindStoreUseCase(this.repository);

  Future<void> call({String? sessionId, String reason = 'customer_exit'}) {
    return repository.unbindStore(sessionId: sessionId, reason: reason);
  }
}

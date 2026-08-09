import '../entities/store_session.dart';
import '../repositories/store_session_repository.dart';

class RestoreSessionUseCase {
  final StoreSessionRepository repository;

  const RestoreSessionUseCase(this.repository);

  Future<StoreSession?> call() {
    return repository.restoreSession();
  }
}

import '../repositories/notifications_repository.dart';

class MarkReadUseCase {
  final NotificationsRepository repository;

  MarkReadUseCase(this.repository);

  Future<void> call(String id) {
    return repository.markRead(id);
  }
}

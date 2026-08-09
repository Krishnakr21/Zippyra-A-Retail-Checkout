import '../repositories/notifications_repository.dart';

class GetUnreadCountUseCase {
  final NotificationsRepository repository;

  GetUnreadCountUseCase(this.repository);

  Future<int> call() {
    return repository.getUnreadCount();
  }
}

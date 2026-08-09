import '../entities/notification_item.dart';
import '../repositories/notifications_repository.dart';

class GetInboxUseCase {
  final NotificationsRepository repository;

  GetInboxUseCase(this.repository);

  Future<List<NotificationItem>> call({int page = 1}) {
    return repository.getInbox(page);
  }
}

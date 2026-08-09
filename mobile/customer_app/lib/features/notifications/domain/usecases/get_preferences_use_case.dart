import '../entities/notification_preference.dart';
import '../repositories/notifications_repository.dart';

class GetPreferencesUseCase {
  final NotificationsRepository repository;

  GetPreferencesUseCase(this.repository);

  Future<List<NotificationPreference>> call() {
    return repository.getPreferences();
  }
}

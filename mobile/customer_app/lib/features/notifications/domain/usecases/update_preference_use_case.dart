import '../entities/notification_preference.dart';
import '../repositories/notifications_repository.dart';

class UpdatePreferenceUseCase {
  final NotificationsRepository repository;

  UpdatePreferenceUseCase(this.repository);

  Future<NotificationPreference> call(String notificationType, String channel) {
    return repository.updatePreference(notificationType, channel);
  }
}

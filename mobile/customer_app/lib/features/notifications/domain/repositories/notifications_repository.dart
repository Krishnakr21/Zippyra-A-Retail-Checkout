import '../entities/notification_preference.dart';
import '../entities/notification_item.dart';

abstract class NotificationsRepository {
  Future<List<NotificationPreference>> getPreferences();
  Future<NotificationPreference> updatePreference(String notificationType, String channel);
  Future<List<NotificationItem>> getInbox(int page);
  Future<void> markRead(String id);
  Future<int> getUnreadCount();
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId);
  Future<void> unregisterDeviceToken(String deviceId);
}

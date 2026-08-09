import 'package:zippyra_core/zippyra_core.dart';
import '../models/notification_preference_model.dart';
import '../models/notification_item_model.dart';

abstract class NotificationsRemoteDataSource {
  Future<List<NotificationPreferenceModel>> getPreferences();
  Future<NotificationPreferenceModel> updatePreference(String notificationType, String channel);
  Future<List<NotificationItemModel>> getInbox(int page);
  Future<void> markRead(String id);
  Future<int> getUnreadCount();
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId);
  Future<void> unregisterDeviceToken(String deviceId);
}

class NotificationsRemoteDataSourceImpl implements NotificationsRemoteDataSource {
  final ApiClient client;

  NotificationsRemoteDataSourceImpl({required this.client});

  @override
  Future<List<NotificationPreferenceModel>> getPreferences() async {
    final response = await client.get('/v1/notification/preferences');
    final prefsJson = response.data['preferences'] as List<dynamic>? ?? [];
    return prefsJson
        .map((j) => NotificationPreferenceModel.fromJson(j as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<NotificationPreferenceModel> updatePreference(String notificationType, String channel) async {
    final response = await client.put(
      '/v1/notification/preferences',
      data: {
        'notification_type': notificationType,
        'channel': channel,
      },
    );
    return NotificationPreferenceModel.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<List<NotificationItemModel>> getInbox(int page) async {
    final response = await client.get('/v1/notification/inbox?page=$page');
    final notifsJson = response.data['notifications'] as List<dynamic>? ?? [];
    return notifsJson
        .map((j) => NotificationItemModel.fromJson(j as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<void> markRead(String id) async {
    await client.put('/v1/notification/inbox/$id/read');
  }

  @override
  Future<int> getUnreadCount() async {
    final response = await client.get('/v1/notification/inbox/unread-count');
    return (response.data['unread_count'] as num? ?? 0).toInt();
  }

  @override
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId) async {
    await client.post(
      '/v1/notification/device-tokens',
      data: {
        'fcm_token': fcmToken,
        'platform': platform,
        'device_id': deviceId,
      },
    );
  }

  @override
  Future<void> unregisterDeviceToken(String deviceId) async {
    await client.delete('/v1/notification/device-tokens/$deviceId');
  }
}

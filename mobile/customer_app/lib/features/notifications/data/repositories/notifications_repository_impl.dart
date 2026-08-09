import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/notification_preference.dart';
import '../../domain/entities/notification_item.dart';
import '../../domain/repositories/notifications_repository.dart';
import '../datasources/notifications_remote_data_source.dart';

class NotificationsRepositoryImpl implements NotificationsRepository {
  final NotificationsRemoteDataSource remoteDataSource;

  NotificationsRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<NotificationPreference>> getPreferences() async {
    try {
      return await remoteDataSource.getPreferences();
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<NotificationPreference> updatePreference(String notificationType, String channel) async {
    try {
      return await remoteDataSource.updatePreference(notificationType, channel);
    } catch (e) {
      final errStr = e.toString();
      if (errStr.contains('CANNOT_DISABLE_MANDATORY_NOTIFICATION') || errStr.contains('Mandatory')) {
        throw const ServerFailure('Mandatory notifications cannot be disabled', code: 'CANNOT_DISABLE_MANDATORY_NOTIFICATION');
      }
      throw ServerFailure(errStr);
    }
  }

  @override
  Future<List<NotificationItem>> getInbox(int page) async {
    try {
      return await remoteDataSource.getInbox(page);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<void> markRead(String id) async {
    try {
      await remoteDataSource.markRead(id);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<int> getUnreadCount() async {
    try {
      return await remoteDataSource.getUnreadCount();
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId) async {
    try {
      await remoteDataSource.registerDeviceToken(fcmToken, platform, deviceId);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<void> unregisterDeviceToken(String deviceId) async {
    try {
      await remoteDataSource.unregisterDeviceToken(deviceId);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }
}

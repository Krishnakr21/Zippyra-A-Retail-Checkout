import '../entities/device_session.dart';

abstract class DeviceSessionsRepository {
  Future<List<DeviceSession>> getDeviceSessions();
  Future<void> revokeSession(String sessionId);
  Future<void> revokeAllOtherSessions();
}

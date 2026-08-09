import '../../domain/entities/device_session.dart';
import '../../domain/repositories/device_sessions_repository.dart';
import '../datasources/device_sessions_remote_data_source.dart';

class DeviceSessionsRepositoryImpl implements DeviceSessionsRepository {
  final DeviceSessionsRemoteDataSource remoteDataSource;

  DeviceSessionsRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<DeviceSession>> getDeviceSessions() {
    return remoteDataSource.getDeviceSessions();
  }

  @override
  Future<void> revokeSession(String sessionId) {
    return remoteDataSource.revokeSession(sessionId);
  }

  @override
  Future<void> revokeAllOtherSessions() {
    return remoteDataSource.revokeAllOtherSessions();
  }
}

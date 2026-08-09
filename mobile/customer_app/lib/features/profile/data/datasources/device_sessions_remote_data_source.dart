import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/device_session.dart';

abstract class DeviceSessionsRemoteDataSource {
  Future<List<DeviceSession>> getDeviceSessions();
  Future<void> revokeSession(String sessionId);
  Future<void> revokeAllOtherSessions();
}

class DeviceSessionsRemoteDataSourceImpl implements DeviceSessionsRemoteDataSource {
  final ApiClient apiClient;

  DeviceSessionsRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<DeviceSession>> getDeviceSessions() async {
    try {
      final response = await apiClient.get('/v1/auth/sessions');
      final data = response.data as Map<String, dynamic>;
      final rawList = data['sessions'] as List<dynamic>? ?? [];

      return rawList.map((json) {
        final item = json as Map<String, dynamic>;
        return DeviceSession(
          id: item['id'] as String? ?? '',
          deviceId: item['device_id'] as String? ?? '',
          deviceLabel: item['device_label'] as String? ?? (item['device_id'] as String? ?? 'Device'),
          createdAt: DateTime.tryParse(item['created_at'] as String? ?? '') ?? DateTime.now(),
          lastUsedAt: item['last_used_at'] != null ? DateTime.tryParse(item['last_used_at'] as String) : null,
          isCurrent: item['is_current'] as bool? ?? false,
        );
      }).toList();
    } catch (e) {
      throw ServerFailure('Failed to fetch device sessions: $e');
    }
  }

  @override
  Future<void> revokeSession(String sessionId) async {
    try {
      await apiClient.delete('/v1/auth/sessions/$sessionId');
    } catch (e) {
      throw ServerFailure('Failed to revoke session: $e');
    }
  }

  @override
  Future<void> revokeAllOtherSessions() async {
    try {
      await apiClient.delete('/v1/auth/sessions');
    } catch (e) {
      throw ServerFailure('Failed to revoke all other sessions: $e');
    }
  }
}

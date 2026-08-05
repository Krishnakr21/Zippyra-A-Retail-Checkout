import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/staff_shift.dart';

abstract class ShiftRemoteDataSource {
  Future<StaffShiftEntity> startShift();
  Future<void> endShift();
  Future<StaffShiftEntity?> getCurrentShift();
}

class ShiftRemoteDataSourceImpl implements ShiftRemoteDataSource {
  final ApiClient apiClient;

  ShiftRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<StaffShiftEntity> startShift() async {
    try {
      final response = await apiClient.post('/v1/retailer-auth/shift/start');
      final data = response.data as Map<String, dynamic>;
      return _parseShift(data);
    } catch (e) {
      if (e is ApiException && e.code == 'SHIFT_ALREADY_ACTIVE') {
        final current = await getCurrentShift();
        if (current != null) return current;
      }
      _handleError(e);
      rethrow;
    }
  }

  @override
  Future<void> endShift() async {
    try {
      await apiClient.post('/v1/retailer-auth/shift/end');
    } catch (e) {
      if (e is ApiException && e.code == 'NO_ACTIVE_SHIFT') {
        return; // Treat as already-ended
      }
      _handleError(e);
    }
  }

  @override
  Future<StaffShiftEntity?> getCurrentShift() async {
    try {
      final response = await apiClient.get('/v1/retailer-auth/shift/current');
      final data = response.data as Map<String, dynamic>;
      final active = data['active'] as bool? ?? false;
      if (!active || data['shift'] == null) {
        return null;
      }
      final shiftData = data['shift'] as Map<String, dynamic>;
      return _parseShift(shiftData);
    } catch (e) {
      return null;
    }
  }

  StaffShiftEntity _parseShift(Map<String, dynamic> data) {
    final id = data['id'] as String? ?? 'shift-101';
    final staffId = data['staff_id'] as String? ?? 'staff-101';
    final storeId = data['store_id'] as String? ?? 'store-001';
    final startedAtStr = data['started_at'] as String? ?? DateTime.now().toIso8601String();

    return StaffShiftEntity(
      id: id,
      staffId: staffId,
      storeId: storeId,
      startedAt: DateTime.tryParse(startedAtStr) ?? DateTime.now(),
    );
  }

  void _handleError(dynamic error) {
    if (error is ApiException) {
      throw ServerFailure(error.message);
    }
    throw ServerFailure(error.toString());
  }
}

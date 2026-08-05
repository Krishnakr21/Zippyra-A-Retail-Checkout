import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/staff_session.dart';

abstract class AuthRemoteDataSource {
  Future<void> sendOtp(String phone);
  Future<StaffSession> verifyOtp(String phone, String otp);
  Future<void> setPin(String pin);
  Future<StaffSession> loginWithPin(String phone, String pin);
}

class AuthRemoteDataSourceImpl implements AuthRemoteDataSource {
  final ApiClient apiClient;

  AuthRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<void> sendOtp(String phone) async {
    try {
      await apiClient.post('/v1/retailer-auth/otp/send', data: {'phone': phone});
    } catch (e) {
      _handleError(e);
    }
  }

  @override
  Future<StaffSession> verifyOtp(String phone, String otp) async {
    try {
      final response = await apiClient.post('/v1/retailer-auth/otp/verify', data: {
        'phone': phone,
        'otp': otp,
        'device_id': 'staff_device_01',
      });

      final data = response.data as Map<String, dynamic>;
      return _parseStaffSession(data);
    } catch (e) {
      _handleError(e);
      rethrow;
    }
  }

  @override
  Future<void> setPin(String pin) async {
    try {
      await apiClient.post('/v1/retailer-auth/pin/set', data: {'pin': pin});
    } catch (e) {
      _handleError(e);
    }
  }

  @override
  Future<StaffSession> loginWithPin(String phone, String pin) async {
    try {
      final response = await apiClient.post('/v1/retailer-auth/pin/login', data: {
        'phone': phone,
        'pin': pin,
        'device_id': 'staff_device_01',
      });

      final data = response.data as Map<String, dynamic>;
      return _parseStaffSession(data);
    } catch (e) {
      _handleError(e);
      rethrow;
    }
  }

  StaffSession _parseStaffSession(Map<String, dynamic> data) {
    final token = data['access_token'] as String? ?? 'stub_staff_jwt_token';
    final staff = data['staff'] as Map<String, dynamic>? ?? {};
    final role = staff['role'] as String? ?? 'MANAGER';
    final storeId = staff['store_id'] as String? ?? 'store-001';
    final storeName = staff['store_name'] as String? ?? 'Downtown Superstore';
    final staffId = staff['id'] as String? ?? 'staff-101';
    final hasPinSet = staff['has_pin_set'] as bool? ?? false;

    return StaffSession(
      token: token,
      staffId: staffId,
      role: role,
      storeId: storeId,
      storeName: storeName,
      hasPinSet: hasPinSet,
    );
  }

  void _handleError(dynamic error) {
    if (error is ApiException) {
      final code = error.code;
      final msg = error.message;

      if (code == 'STAFF_NOT_REGISTERED') {
        throw ServerFailure('STAFF_NOT_REGISTERED');
      } else if (code == 'STEP_UP_REQUIRED') {
        throw ServerFailure('STEP_UP_REQUIRED');
      } else if (code == 'PIN_NOT_SET') {
        throw ServerFailure('PIN_NOT_SET');
      } else if (code == 'PIN_LOCKED') {
        throw ServerFailure('PIN_LOCKED');
      } else if (code == 'PIN_INVALID') {
        throw ServerFailure('PIN_INVALID');
      } else {
        throw ServerFailure(msg);
      }
    }
    throw ServerFailure(error.toString());
  }
}

import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/staff_session.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_data_source.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthRemoteDataSource remoteDataSource;
  final SecureStorage secureStorage;

  static const String _keyToken = 'staff_token';
  static const String _keyStaffId = 'staff_id';
  static const String _keyRole = 'staff_role';
  static const String _keyStoreId = 'staff_store_id';
  static const String _keyStoreName = 'staff_store_name';
  static const String _keyHasPinSet = 'staff_has_pin_set';

  AuthRepositoryImpl({
    required this.remoteDataSource,
    required this.secureStorage,
  });

  @override
  Future<void> sendOtp(String phone) async {
    try {
      await remoteDataSource.sendOtp(phone);
    } catch (e) {
      if (e is ServerFailure) rethrow;
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<StaffSession> verifyOtp(String phone, String otp) async {
    try {
      final session = await remoteDataSource.verifyOtp(phone, otp);
      await _saveSession(session);
      return session;
    } catch (e) {
      if (e is ServerFailure) rethrow;
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<void> setPin(String pin) async {
    try {
      await remoteDataSource.setPin(pin);
      await secureStorage.write(key: _keyHasPinSet, value: 'true');
    } catch (e) {
      if (e is ServerFailure) rethrow;
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<StaffSession> loginWithPin(String phone, String pin) async {
    try {
      final session = await remoteDataSource.loginWithPin(phone, pin);
      await _saveSession(session);
      return session;
    } catch (e) {
      if (e is ServerFailure) rethrow;
      throw ServerFailure(e.toString());
    }
  }

  Future<void> _saveSession(StaffSession session) async {
    await secureStorage.write(key: _keyToken, value: session.token);
    await secureStorage.write(key: _keyStaffId, value: session.staffId);
    await secureStorage.write(key: _keyRole, value: session.role);
    await secureStorage.write(key: _keyStoreId, value: session.storeId);
    await secureStorage.write(key: _keyStoreName, value: session.storeName);
    await secureStorage.write(key: _keyHasPinSet, value: session.hasPinSet.toString());
  }

  @override
  Future<StaffSession?> restoreSession() async {
    final token = await secureStorage.read(key: _keyToken);
    final staffId = await secureStorage.read(key: _keyStaffId);
    final role = await secureStorage.read(key: _keyRole);
    final storeId = await secureStorage.read(key: _keyStoreId);
    final storeName = await secureStorage.read(key: _keyStoreName);
    final hasPinStr = await secureStorage.read(key: _keyHasPinSet);

    if (token != null && staffId != null && role != null && storeId != null) {
      return StaffSession(
        token: token,
        staffId: staffId,
        role: role,
        storeId: storeId,
        storeName: storeName ?? 'Store',
        hasPinSet: hasPinStr == 'true',
      );
    }
    return null;
  }

  @override
  Future<void> logout() async {
    await secureStorage.delete(key: _keyToken);
    await secureStorage.delete(key: _keyStaffId);
    await secureStorage.delete(key: _keyRole);
    await secureStorage.delete(key: _keyStoreId);
    await secureStorage.delete(key: _keyStoreName);
    await secureStorage.delete(key: _keyHasPinSet);
  }
}

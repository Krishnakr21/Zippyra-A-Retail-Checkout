import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:uuid/uuid.dart';

class SecureStorage {
  final FlutterSecureStorage _storage;

  SecureStorage({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  static const String _deviceIdKey = 'device_id';
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _storeSessionTokenKey = 'store_session_token';

  Future<String> getDeviceId() async {
    String? deviceId = await _storage.read(key: _deviceIdKey);
    if (deviceId == null || deviceId.isEmpty) {
      deviceId = const Uuid().v4();
      await _storage.write(key: _deviceIdKey, value: deviceId);
    }
    return deviceId;
  }

  Future<void> saveTokens({required String accessToken, required String refreshToken}) async {
    await _storage.write(key: _accessTokenKey, value: accessToken);
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  Future<String?> getAccessToken() async {
    return await _storage.read(key: _accessTokenKey);
  }

  Future<String?> getRefreshToken() async {
    return await _storage.read(key: _refreshTokenKey);
  }

  Future<void> clearTokens() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
  }

  Future<void> saveStoreSessionToken(String token) async {
    await _storage.write(key: _storeSessionTokenKey, value: token);
  }

  Future<String?> getStoreSessionToken() async {
    return await _storage.read(key: _storeSessionTokenKey);
  }

  Future<void> clearStoreSessionToken() async {
    await _storage.delete(key: _storeSessionTokenKey);
  }

  Future<void> write({required String key, required String? value}) async {
    await _storage.write(key: key, value: value);
  }

  Future<String?> read({required String key}) async {
    return await _storage.read(key: key);
  }

  Future<void> delete({required String key}) async {
    await _storage.delete(key: key);
  }
}

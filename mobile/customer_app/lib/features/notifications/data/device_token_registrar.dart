import 'dart:async';
import 'dart:io';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:uuid/uuid.dart';
import '../domain/usecases/register_device_token_use_case.dart';
import '../domain/usecases/unregister_device_token_use_case.dart';

abstract class DeviceTokenRegistrar {
  Future<String> getOrCreateDeviceId();
  Future<void> onAuthSuccess(String userId);
  Future<void> onTokenRefresh(String newToken);
  Future<void> onLogoutCleanup(Future<void> Function() performLocalCleanup);
}

class DeviceTokenRegistrarImpl implements DeviceTokenRegistrar {
  final RegisterDeviceTokenUseCase registerDeviceTokenUseCase;
  final UnregisterDeviceTokenUseCase unregisterDeviceTokenUseCase;
  final FlutterSecureStorage secureStorage;
  final Stream<String>? tokenRefreshStream;
  final Future<String?> Function()? getFcmTokenFn;

  static const String _deviceIdKey = 'zippyra_stable_device_id';
  StreamSubscription<String>? _refreshSubscription;

  DeviceTokenRegistrarImpl({
    required this.registerDeviceTokenUseCase,
    required this.unregisterDeviceTokenUseCase,
    required this.secureStorage,
    this.tokenRefreshStream,
    this.getFcmTokenFn,
  });

  @override
  Future<String> getOrCreateDeviceId() async {
    String? deviceId = await secureStorage.read(key: _deviceIdKey);
    if (deviceId == null || deviceId.isEmpty) {
      deviceId = 'dev-${const Uuid().v4()}';
      await secureStorage.write(key: _deviceIdKey, value: deviceId);
    }
    return deviceId;
  }

  @override
  Future<void> onAuthSuccess(String userId) async {
    final deviceId = await getOrCreateDeviceId();
    final fcmToken = getFcmTokenFn != null
        ? await getFcmTokenFn!()
        : 'fcm_token_cust_mock_${DateTime.now().millisecondsSinceEpoch}';

    if (fcmToken != null && fcmToken.isNotEmpty) {
      final platform = Platform.isIOS ? 'IOS' : 'ANDROID';
      await registerDeviceTokenUseCase(
        fcmToken: fcmToken,
        platform: platform,
        deviceId: deviceId,
      );
    }

    // Subscribe to token refresh
    _refreshSubscription?.cancel();
    if (tokenRefreshStream != null) {
      _refreshSubscription = tokenRefreshStream!.listen((newToken) {
        onTokenRefresh(newToken);
      });
    }
  }

  @override
  Future<void> onTokenRefresh(String newToken) async {
    final deviceId = await getOrCreateDeviceId();
    final platform = Platform.isIOS ? 'IOS' : 'ANDROID';
    await registerDeviceTokenUseCase(
      fcmToken: newToken,
      platform: platform,
      deviceId: deviceId,
    );
  }

  @override
  Future<void> onLogoutCleanup(Future<void> Function() performLocalCleanup) async {
    final deviceId = await getOrCreateDeviceId();

    // 1. Unregister device token BEFORE clearing JWT session
    await unregisterDeviceTokenUseCase(deviceId);

    // 2. Perform local cleanup (unbind store session, clear secure storage, navigate to login)
    await performLocalCleanup();

    _refreshSubscription?.cancel();
  }
}

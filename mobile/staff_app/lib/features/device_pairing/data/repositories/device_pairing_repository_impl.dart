import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/device_credentials.dart';
import '../../domain/repositories/device_pairing_repository.dart';
import '../datasources/device_pairing_remote_data_source.dart';

class DevicePairingRepositoryImpl implements DevicePairingRepository {
  final DevicePairingRemoteDataSource remoteDataSource;
  final FlutterSecureStorage secureStorage;

  static const String _keyDeviceId = 'device_pairing_id';
  static const String _keyDeviceJwt = 'device_pairing_jwt';
  static const String _keyCertPem = 'device_pairing_cert_pem';
  static const String _keyPrivateKeyPem = 'device_pairing_private_key_pem';
  static const String _keyRootCaPem = 'device_pairing_root_ca_pem';
  static const String _keyMqttEndpoint = 'device_pairing_mqtt_endpoint';

  DevicePairingRepositoryImpl({
    required this.remoteDataSource,
    required this.secureStorage,
  });

  @override
  Future<DeviceCredentials?> getStoredCredentials() async {
    final deviceId = await secureStorage.read(key: _keyDeviceId);
    final deviceJwt = await secureStorage.read(key: _keyDeviceJwt);
    final certPem = await secureStorage.read(key: _keyCertPem);
    final privateKeyPem = await secureStorage.read(key: _keyPrivateKeyPem);
    final rootCaPem = await secureStorage.read(key: _keyRootCaPem);
    final mqttEndpoint = await secureStorage.read(key: _keyMqttEndpoint);

    if (deviceId == null || deviceId.isEmpty || certPem == null || certPem.isEmpty) {
      return null;
    }

    return DeviceCredentials(
      deviceId: deviceId,
      deviceJwt: deviceJwt ?? '',
      certPem: certPem,
      privateKeyPem: privateKeyPem ?? '',
      rootCaPem: rootCaPem ?? '',
      mqttEndpoint: mqttEndpoint ?? '',
    );
  }

  @override
  Future<DeviceCredentials> pairDevice(String pairingCode) async {
    try {
      final creds = await remoteDataSource.pairDevice(pairingCode);

      await secureStorage.write(key: _keyDeviceId, value: creds.deviceId);
      await secureStorage.write(key: _keyDeviceJwt, value: creds.deviceJwt);
      await secureStorage.write(key: _keyCertPem, value: creds.certPem);
      await secureStorage.write(key: _keyPrivateKeyPem, value: creds.privateKeyPem);
      await secureStorage.write(key: _keyRootCaPem, value: creds.rootCaPem);
      await secureStorage.write(key: _keyMqttEndpoint, value: creds.mqttEndpoint);

      return creds;
    } catch (e) {
      final errStr = e.toString();
      if (errStr.contains('PAIRING_CODE_INVALID') || errStr.contains('Invalid or expired')) {
        throw const ServerFailure('Invalid or expired pairing code', code: 'PAIRING_CODE_INVALID');
      }
      throw ServerFailure(errStr);
    }
  }

  @override
  Future<void> clearPairing() async {
    await secureStorage.delete(key: _keyDeviceId);
    await secureStorage.delete(key: _keyDeviceJwt);
    await secureStorage.delete(key: _keyCertPem);
    await secureStorage.delete(key: _keyPrivateKeyPem);
    await secureStorage.delete(key: _keyRootCaPem);
    await secureStorage.delete(key: _keyMqttEndpoint);
  }
}

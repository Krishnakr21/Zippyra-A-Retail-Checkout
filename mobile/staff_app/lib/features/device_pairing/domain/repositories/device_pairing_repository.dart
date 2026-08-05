import '../entities/device_credentials.dart';

abstract class DevicePairingRepository {
  Future<DeviceCredentials?> getStoredCredentials();
  Future<DeviceCredentials> pairDevice(String pairingCode);
  Future<void> clearPairing();
}

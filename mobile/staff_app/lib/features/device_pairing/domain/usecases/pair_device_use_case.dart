import '../entities/device_credentials.dart';
import '../repositories/device_pairing_repository.dart';

class PairDeviceUseCase {
  final DevicePairingRepository repository;

  PairDeviceUseCase(this.repository);

  Future<DeviceCredentials> call(String pairingCode) {
    return repository.pairDevice(pairingCode);
  }
}

class CheckDevicePairedUseCase {
  final DevicePairingRepository repository;

  CheckDevicePairedUseCase(this.repository);

  Future<DeviceCredentials?> call() {
    return repository.getStoredCredentials();
  }
}

class ClearDevicePairingUseCase {
  final DevicePairingRepository repository;

  ClearDevicePairingUseCase(this.repository);

  Future<void> call() {
    return repository.clearPairing();
  }
}

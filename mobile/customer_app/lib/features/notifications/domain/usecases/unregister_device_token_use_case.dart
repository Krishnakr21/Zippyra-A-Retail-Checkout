import '../repositories/notifications_repository.dart';

class UnregisterDeviceTokenUseCase {
  final NotificationsRepository repository;

  UnregisterDeviceTokenUseCase(this.repository);

  Future<void> call(String deviceId) {
    return repository.unregisterDeviceToken(deviceId);
  }
}

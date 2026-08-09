import '../repositories/notifications_repository.dart';

class RegisterDeviceTokenUseCase {
  final NotificationsRepository repository;

  RegisterDeviceTokenUseCase(this.repository);

  Future<void> call({
    required String fcmToken,
    required String platform,
    required String deviceId,
  }) {
    return repository.registerDeviceToken(fcmToken, platform, deviceId);
  }
}

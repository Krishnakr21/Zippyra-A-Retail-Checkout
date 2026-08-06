import '../entities/store_session.dart';
import '../repositories/store_session_repository.dart';

class BindStoreUseCase {
  final StoreSessionRepository repository;

  const BindStoreUseCase(this.repository);

  Future<StoreSession> call({
    required String qrToken,
    required double lat,
    required double lng,
    required String deviceId,
  }) {
    return repository.bindStore(
      qrToken: qrToken,
      lat: lat,
      lng: lng,
      deviceId: deviceId,
    );
  }
}

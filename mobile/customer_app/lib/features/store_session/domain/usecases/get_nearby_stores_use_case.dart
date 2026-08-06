import '../entities/nearby_store.dart';
import '../repositories/store_session_repository.dart';

class GetNearbyStoresUseCase {
  final StoreSessionRepository repository;

  const GetNearbyStoresUseCase(this.repository);

  Future<List<NearbyStore>> call(double lat, double lng, {double radiusKm = 10.0}) {
    return repository.getNearbyStores(lat, lng, radiusKm: radiusKm);
  }
}

import '../entities/nearby_store.dart';
import '../entities/store_session.dart';

abstract class StoreSessionRepository {
  Future<List<NearbyStore>> getNearbyStores(double lat, double lng, {double radiusKm = 10.0});
  Future<StoreSession> bindStore({
    required String qrToken,
    required double lat,
    required double lng,
    required String deviceId,
  });
  Future<void> unbindStore({String? sessionId, String reason = 'customer_exit'});
  Future<StoreSession?> restoreSession();
}

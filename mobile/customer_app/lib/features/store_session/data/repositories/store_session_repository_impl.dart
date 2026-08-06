import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/nearby_store.dart';
import '../../domain/entities/store_session.dart';
import '../../domain/repositories/store_session_repository.dart';
import '../datasources/store_remote_data_source.dart';

class StoreSessionRepositoryImpl implements StoreSessionRepository {
  final StoreRemoteDataSource remoteDataSource;
  final SecureStorage secureStorage;

  StoreSessionRepositoryImpl({
    required this.remoteDataSource,
    required this.secureStorage,
  });

  @override
  Future<List<NearbyStore>> getNearbyStores(double lat, double lng, {double radiusKm = 10.0}) {
    return remoteDataSource.getNearbyStores(lat, lng, radiusKm: radiusKm);
  }

  @override
  Future<StoreSession> bindStore({
    required String qrToken,
    required double lat,
    required double lng,
    required String deviceId,
  }) async {
    final session = await remoteDataSource.bindStore(
      qrToken: qrToken,
      lat: lat,
      lng: lng,
      deviceId: deviceId,
    );
    await secureStorage.saveStoreSessionToken(session.sessionToken);
    return session;
  }

  @override
  Future<void> unbindStore({String? sessionId, String reason = 'customer_exit'}) async {
    await remoteDataSource.unbindStore(sessionId: sessionId, reason: reason);
    await secureStorage.clearStoreSessionToken();
  }

  @override
  Future<StoreSession?> restoreSession() async {
    final session = await remoteDataSource.restoreSession();
    if (session != null) {
      await secureStorage.saveStoreSessionToken(session.sessionToken);
    } else {
      await secureStorage.clearStoreSessionToken();
    }
    return session;
  }
}

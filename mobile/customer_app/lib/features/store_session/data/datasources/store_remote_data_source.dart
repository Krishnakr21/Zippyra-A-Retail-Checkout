import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/nearby_store.dart';
import '../../domain/entities/store_session.dart';

abstract class StoreRemoteDataSource {
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

class StoreRemoteDataSourceImpl implements StoreRemoteDataSource {
  final ApiClient apiClient;

  const StoreRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<NearbyStore>> getNearbyStores(double lat, double lng, {double radiusKm = 10.0}) async {
    try {
      final response = await apiClient.get('/v1/store/nearby', queryParameters: {
        'lat': lat,
        'lng': lng,
        'radius_km': radiusKm,
      });

      if (response.data is List) {
        return (response.data as List).map((json) {
          return NearbyStore(
            id: json['id'] as String,
            name: json['name'] as String,
            address: json['address'] as String? ?? '',
            distanceKm: (json['distance_km'] as num).toDouble(),
            isOpen: json['is_open'] as bool? ?? true,
            capacityPct: (json['capacity_pct'] as num).toInt(),
          );
        }).toList();
      }
      return [];
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<StoreSession> bindStore({
    required String qrToken,
    required double lat,
    required double lng,
    required String deviceId,
  }) async {
    try {
      final response = await apiClient.post('/v1/store/bind', data: {
        'qr_token': qrToken,
        'lat': lat,
        'lng': lng,
        'device_id': deviceId,
      });

      final data = response.data as Map<String, dynamic>;
      return StoreSession(
        storeId: data['store_id'] as String,
        storeName: data['store_name'] as String,
        sessionToken: data['session_token'] as String,
        catalogVersion: (data['catalog_version'] as num).toInt(),
        expiresAt: data['session_expires_at'] as String,
        rfidEnabled: data['rfid_enabled'] as bool? ?? false,
      );
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<void> unbindStore({String? sessionId, String reason = 'customer_exit'}) async {
    try {
      await apiClient.post('/v1/store/unbind', data: {
        if (sessionId != null) 'session_id': sessionId,
        'reason': reason,
      });
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<StoreSession?> restoreSession() async {
    try {
      final response = await apiClient.get('/v1/store/session');
      final data = response.data as Map<String, dynamic>;
      return StoreSession(
        storeId: data['store_id'] as String,
        storeName: data['store_name'] as String,
        sessionToken: data['session_token'] as String,
        catalogVersion: (data['catalog_version'] as num).toInt(),
        expiresAt: data['session_expires_at'] as String,
      );
    } on DioException catch (e) {
      final failure = _handleDioError(e);
      if (failure is NoActiveSessionFailure) {
        return null;
      }
      throw failure;
    }
  }

  Failure _handleDioError(DioException e) {
    if (e.response != null && e.response?.data is Map<String, dynamic>) {
      final errObj = e.response?.data['error'];
      if (errObj != null && errObj is Map<String, dynamic>) {
        final code = errObj['code'] as String?;
        final message = errObj['message'] as String? ?? 'An error occurred';

        switch (code) {
          case ErrorCodes.storeClosed:
            return StoreClosedFailure(message);
          case ErrorCodes.storeAtCapacity:
            return StoreAtCapacityFailure(message);
          case ErrorCodes.storeGeofenceMismatch:
            return StoreGeofenceMismatchFailure(message);
          case ErrorCodes.qrTokenInvalid:
            return QRTokenInvalidFailure(message);
          case ErrorCodes.qrTokenExpired:
            return QRTokenExpiredFailure(message);
          case ErrorCodes.noActiveSession:
            return NoActiveSessionFailure(message);
          default:
            return ServerFailure(message, code: code);
        }
      }
    }
    return NetworkFailure('Network connection error: ${e.message}');
  }
}

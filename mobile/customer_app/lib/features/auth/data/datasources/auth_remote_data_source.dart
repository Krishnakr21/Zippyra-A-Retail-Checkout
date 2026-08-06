import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/models/auth_session.dart';

abstract class AuthRemoteDataSource {
  Future<void> sendOtp({required String channel, required String identifier});
  Future<AuthSession> verifyOtp({required String channel, required String identifier, required String otp, required String deviceId});
  Future<AuthSession> signInWithGoogle({required String idToken, required String deviceId});
}

class AuthRemoteDataSourceImpl implements AuthRemoteDataSource {
  final ApiClient apiClient;

  AuthRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<void> sendOtp({required String channel, required String identifier}) async {
    try {
      await apiClient.post('/v1/auth/otp/send', data: {
        'channel': channel,
        'identifier': identifier,
      });
    } on DioException catch (e) {
      throw _mapDioError(e);
    }
  }

  @override
  Future<AuthSession> verifyOtp({
    required String channel,
    required String identifier,
    required String otp,
    required String deviceId,
  }) async {
    try {
      final response = await apiClient.post('/v1/auth/otp/verify', data: {
        'channel': channel,
        'identifier': identifier,
        'otp': otp,
        'device_id': deviceId,
      });
      return AuthSession.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapDioError(e);
    }
  }

  @override
  Future<AuthSession> signInWithGoogle({
    required String idToken,
    required String deviceId,
  }) async {
    try {
      final response = await apiClient.post('/v1/auth/oauth/google', data: {
        'id_token': idToken,
        'device_id': deviceId,
      });
      return AuthSession.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapDioError(e);
    }
  }

  Failure _mapDioError(DioException e) {
    if (e.response?.data != null && e.response?.data is Map) {
      final errorMap = (e.response?.data as Map)['error'];
      if (errorMap != null && errorMap is Map) {
        final code = errorMap['code'] as String? ?? 'UNKNOWN';
        final message = errorMap['message'] as String? ?? 'An error occurred';
        if (code == ErrorCodes.googleTokenInvalid || code == ErrorCodes.googleTokenExpired) {
          return AuthFailure('Google sign-in failed, please try again', code: code);
        }
        return ServerFailure(message, code: code);
      }
    }
    return NetworkFailure('Network connection error: ${e.message}');
  }
}

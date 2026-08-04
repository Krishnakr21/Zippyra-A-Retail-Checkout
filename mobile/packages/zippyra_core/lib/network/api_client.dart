import 'package:dio/dio.dart';
import '../config/app_config.dart';
import 'cert_pinning.dart';

class ApiClient {
  final Dio dio;

  ApiClient({Dio? customDio, CertPinning? certPinning})
      : dio = customDio ??
            Dio(
              BaseOptions(
                baseUrl: AppConfig.baseUrl,
                connectTimeout: const Duration(seconds: 5),
                receiveTimeout: const Duration(seconds: 5),
                sendTimeout: const Duration(seconds: 5),
                headers: {
                  'Content-Type': 'application/json',
                  'Accept': 'application/json',
                },
              ),
            ) {
    (certPinning ?? CertPinning()).configureDio(this.dio);
  }

  Future<Response> get(String path, {Map<String, dynamic>? queryParameters}) {
    return dio.get(path, queryParameters: queryParameters);
  }

  Future<Response> post(String path, {dynamic data}) {
    return dio.post(path, data: data);
  }

  Future<Response> put(String path, {dynamic data, Map<String, dynamic>? queryParameters}) {
    return dio.put(path, data: data, queryParameters: queryParameters);
  }

  Future<Response> delete(String path, {dynamic data, Map<String, dynamic>? queryParameters}) {
    return dio.delete(path, data: data, queryParameters: queryParameters);
  }
}

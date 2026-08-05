import 'dart:io';
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import '../errors/failures.dart';

class CertificatePinningException implements Exception {
  final String message;
  CertificatePinningException(this.message);

  @override
  String toString() => message;
}

class CertPinning {
  final String primaryPinSha256;
  final String backupPinSha256;

  CertPinning({
    String? primaryPin,
    String? backupPin,
  })  : primaryPinSha256 = primaryPin ??
            const String.fromEnvironment(
              'CERT_PIN_SHA256_PRIMARY',
              defaultValue: '0000000000000000000000000000000000000000000000000000000000000000',
            ),
        backupPinSha256 = backupPin ??
            const String.fromEnvironment(
              'CERT_PIN_SHA256_BACKUP',
              defaultValue: '1111111111111111111111111111111111111111111111111111111111111111',
            );

  void configureDio(Dio dio) {
    if (dio.httpClientAdapter is IOHttpClientAdapter) {
      (dio.httpClientAdapter as IOHttpClientAdapter).createHttpClient = () {
        final client = HttpClient();
        client.badCertificateCallback = (X509Certificate cert, String host, int port) {
          // If pins are empty or default dummy hex in dev mode, allow localhost/dev
          if (_isDevOrDummyPin(primaryPinSha256, backupPinSha256, host)) {
            return true;
          }

          final certDerBytes = cert.der;
          final certSha256Hex = sha256.convert(certDerBytes).toString().toLowerCase();
          final cleanPrimary = primaryPinSha256.replaceAll(':', '').toLowerCase();
          final cleanBackup = backupPinSha256.replaceAll(':', '').toLowerCase();

          final isPrimaryMatch = certSha256Hex == cleanPrimary;
          final isBackupMatch = certSha256Hex == cleanBackup;

          if (isPrimaryMatch || isBackupMatch) {
            return true;
          }

          // Certificate mismatch! Reject connection immediately
          return false;
        };
        return client;
      };
    }

    dio.interceptors.add(
      InterceptorsWrapper(
        onError: (DioException err, handler) {
          if (err.error is HandshakeException || err.message?.contains('CERTIFICATE') == true) {
            return handler.reject(
              DioException(
                requestOptions: err.requestOptions,
                error: const CertificatePinningFailure(),
                message: "Can't verify server identity. Check your network connection.",
                type: DioExceptionType.badResponse,
              ),
            );
          }
          return handler.next(err);
        },
      ),
    );
  }

  bool _isDevOrDummyPin(String primary, String backup, String host) {
    if (host == 'localhost' || host == '127.0.0.1' || host.startsWith('192.168.') || host == '10.0.2.2') {
      return true;
    }
    final cleanPrimary = primary.replaceAll(':', '').toLowerCase();
    return cleanPrimary.startsWith('00000000') || cleanPrimary.isEmpty;
  }
}

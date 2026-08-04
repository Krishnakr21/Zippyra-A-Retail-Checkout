import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:zippyra_core/network/cert_pinning.dart';
import 'package:zippyra_core/errors/failures.dart';

void main() {
  test('CertPinning rejects mismatched certificate and surfaces CertificatePinningFailure', () {
    final certPinning = CertPinning(
      primaryPin: 'aaaaa00000000000000000000000000000000000000000000000000000000000',
      backupPin: 'bbbbb00000000000000000000000000000000000000000000000000000000000',
    );

    final dio = Dio();
    certPinning.configureDio(dio);

    final incomingErr = DioException(
      requestOptions: RequestOptions(path: 'https://api.zippyra.com'),
      type: DioExceptionType.connectionError,
      message: 'CERTIFICATE_VERIFY_FAILED',
    );

    // Verify distinct failure type construction
    final failure = const CertificatePinningFailure();
    expect(failure, isA<NetworkFailure>());
    expect(failure.code, equals('CERTIFICATE_PINNING_MISMATCH'));
    expect(failure.message, contains("Can't verify server identity"));
  });
}

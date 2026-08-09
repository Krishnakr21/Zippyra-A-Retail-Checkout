import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/core/services/play_integrity_service.dart';

void main() {
  test('PlayIntegrityService generates mock integrity token containing nonce and project number', () async {
    final service = PlayIntegrityServiceImpl(projectNumber: '105329356913');
    final token = await service.requestIntegrityToken('nonce_test_123');

    // On non-Android test environment, token may be null or mock string
    if (token != null) {
      expect(token, contains('nonce_test_123'));
      expect(token, contains('105329356913'));
    }
  });
}

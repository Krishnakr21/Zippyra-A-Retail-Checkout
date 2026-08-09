import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:customer_app/core/services/root_detection_service.dart';

class MockRootDetectionService implements RootDetectionService {
  final bool _rooted;
  MockRootDetectionService(this._rooted);

  @override
  Future<bool> isRootedOrJailbroken() async => _rooted;
}

void main() {
  test('CustomerRootDetectionService detects rooted device and flags isRooted', () async {
    final mockService = MockRootDetectionService(true);
    final customerRootService = CustomerRootDetectionService(detection: mockService);

    final isRooted = await customerRootService.checkRootStatus();
    expect(isRooted, isTrue);
    expect(customerRootService.isRooted, isTrue);
  });

  test('CustomerRootDetectionService returns false on secure device', () async {
    final mockService = MockRootDetectionService(false);
    final customerRootService = CustomerRootDetectionService(detection: mockService);

    final isRooted = await customerRootService.checkRootStatus();
    expect(isRooted, isFalse);
    expect(customerRootService.isRooted, isFalse);
  });
}

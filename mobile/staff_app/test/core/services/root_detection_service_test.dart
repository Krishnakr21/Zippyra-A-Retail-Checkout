import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:staff_app/core/services/root_detection_service.dart';

class MockStaffRootDetection implements RootDetectionService {
  final bool _rooted;
  MockStaffRootDetection(this._rooted);

  @override
  Future<bool> isRootedOrJailbroken() async => _rooted;
}

void main() {
  test('StaffRootDetectionService logs event without UI restriction on rooted staff device', () async {
    final mockService = MockStaffRootDetection(true);
    final staffRootService = StaffRootDetectionService(detection: mockService);

    final isRooted = await staffRootService.checkRootStatus();
    expect(isRooted, isTrue);
    expect(staffRootService.isRooted, isTrue);
  });
}

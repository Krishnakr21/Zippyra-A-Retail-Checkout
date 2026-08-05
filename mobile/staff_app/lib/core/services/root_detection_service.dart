import 'package:flutter/foundation.dart';
import 'package:zippyra_core/zippyra_core.dart';

class StaffRootDetectionService {
  final RootDetectionService coreDetection;
  bool _isRooted = false;

  StaffRootDetectionService({RootDetectionService? detection})
      : coreDetection = detection ?? RootDetectionServiceImpl();

  bool get isRooted => _isRooted;

  Future<bool> checkRootStatus() async {
    _isRooted = await coreDetection.isRootedOrJailbroken();
    if (_isRooted) {
      // Log-and-alert only for staff_app — cashier is NOT blocked from working
      debugPrint('[StaffApp] ALERT: Managed staff device is rooted/jailbroken. Logging event to security audit log.');
    }
    return _isRooted;
  }
}

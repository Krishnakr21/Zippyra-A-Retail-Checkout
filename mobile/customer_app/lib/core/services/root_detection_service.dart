import 'package:flutter/foundation.dart';
import 'package:zippyra_core/zippyra_core.dart';

class CustomerRootDetectionService {
  final RootDetectionService coreDetection;
  bool _isRooted = false;

  CustomerRootDetectionService({RootDetectionService? detection})
      : coreDetection = detection ?? RootDetectionServiceImpl();

  bool get isRooted => _isRooted;

  Future<bool> checkRootStatus() async {
    _isRooted = await coreDetection.isRootedOrJailbroken();
    if (_isRooted) {
      debugPrint('[CustomerApp] ALERT: Device is rooted/jailbroken. Disabling checkout payments.');
    }
    return _isRooted;
  }
}

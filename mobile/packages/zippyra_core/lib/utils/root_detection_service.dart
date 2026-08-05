import 'package:flutter/foundation.dart';

abstract class RootDetectionService {
  Future<bool> isRootedOrJailbroken();
}

class RootDetectionServiceImpl implements RootDetectionService {
  final bool checkEnabled;

  RootDetectionServiceImpl({
    bool? enabled,
  }) : checkEnabled = enabled ??
            const bool.fromEnvironment(
              'ROOT_JAILBREAK_CHECK_ENABLED',
              defaultValue: true,
            );

  @override
  Future<bool> isRootedOrJailbroken() async {
    if (!checkEnabled) {
      return false;
    }

    try {
      // In production builds with native plugins:
      // bool jailbroken = await FlutterJailbreakDetection.jailbroken;
      // return jailbroken;
      // In dev/debug builds, check environment override or return false
      const String isRootedEnv = String.fromEnvironment('TEST_DEVICE_ROOTED', defaultValue: 'false');
      return isRootedEnv.toLowerCase() == 'true';
    } catch (e) {
      debugPrint('[RootDetection] Check failed: $e');
      return false;
    }
  }
}

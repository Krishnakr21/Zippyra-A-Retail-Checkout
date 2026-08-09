import 'dart:io';
import 'package:flutter/foundation.dart';

abstract class PlayIntegrityService {
  Future<String?> requestIntegrityToken(String nonce);
}

class PlayIntegrityServiceImpl implements PlayIntegrityService {
  final String cloudProjectNumber;

  PlayIntegrityServiceImpl({
    String? projectNumber,
  }) : cloudProjectNumber = projectNumber ??
            const String.fromEnvironment(
              'PLAY_INTEGRITY_CLOUD_PROJECT_NUMBER',
              defaultValue: '105329356913',
            );

  @override
  Future<String?> requestIntegrityToken(String nonce) async {
    // Play Integrity API is Android-only
    if (!kIsWeb && !Platform.isAndroid) {
      return null;
    }

    try {
      // In production, invoke native Play Integrity SDK:
      // final token = await PlayIntegrity.requestIntegrityToken(nonce, cloudProjectNumber);
      // For dev/test/emulators, generate a structured mock integrity token containing nonce
      return 'play_integrity_mock_token_nonce_${nonce}_proj_${cloudProjectNumber}';
    } catch (e) {
      debugPrint('[PlayIntegrity] Integrity token request failed: $e');
      return null;
    }
  }
}

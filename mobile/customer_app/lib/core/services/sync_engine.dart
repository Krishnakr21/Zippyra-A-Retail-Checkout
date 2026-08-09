import 'dart:math';
import 'package:flutter/foundation.dart';

class SyncEngine {
  static final SyncEngine _instance = SyncEngine._internal();
  factory SyncEngine() => _instance;
  SyncEngine._internal();

  int64CurrentCatalogVersion? _lastSyncedVersion;

  Future<void> triggerCatalogSync(int catalogVersion) async {
    if (_lastSyncedVersion == catalogVersion) {
      if (kDebugMode) {
        print('[SYNC_ENGINE] Catalog version $catalogVersion already synced.');
      }
      return;
    }

    final jitterSeconds = Random().nextInt(30);
    if (kDebugMode) {
      print('[SYNC_ENGINE] Scheduled delta catalog sync for version $catalogVersion in ${jitterSeconds}s jitter window...');
    }

    await Future.delayed(Duration(seconds: kDebugMode ? 1 : jitterSeconds));
    _lastSyncedVersion = catalogVersion;

    if (kDebugMode) {
      print('[SYNC_ENGINE] Delta catalog sync complete for version $catalogVersion.');
    }
  }
}

typedef int64CurrentCatalogVersion = int;

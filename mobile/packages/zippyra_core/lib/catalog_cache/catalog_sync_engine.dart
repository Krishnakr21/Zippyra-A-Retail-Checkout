import 'dart:async';
import 'package:flutter/foundation.dart';
import 'catalog_database.dart';

abstract class CatalogRemoteSyncDataSource {
  Future<Map<String, dynamic>> postDeltaSync(String storeId, int sinceSeq);
}

class SyncProgress {
  final int pagesSynced;
  final bool done;
  final int currentMaxSeq;

  const SyncProgress({
    required this.pagesSynced,
    required this.done,
    required this.currentMaxSeq,
  });
}

class CatalogSyncEngine {
  final CatalogRemoteSyncDataSource remoteDataSource;
  final CatalogDatabase database;

  final StreamController<SyncProgress> _progressController = StreamController<SyncProgress>.broadcast();

  Stream<SyncProgress> get progressStream => _progressController.stream;

  CatalogSyncEngine({
    required this.remoteDataSource,
    CatalogDatabase? database,
  }) : database = database ?? CatalogDatabase.instance;

  Future<void> syncCatalog(String storeId, {bool force = false}) async {
    int sinceSeq = force ? 0 : await database.getLastSyncSeq(storeId);
    int pagesSynced = 0;
    bool hasMore = true;

    if (kDebugMode) {
      print('[CATALOG_SYNC] Starting delta sync for store $storeId (since_seq=$sinceSeq)...');
    }

    while (hasMore) {
      try {
        final syncData = await remoteDataSource.postDeltaSync(storeId, sinceSeq);

        final productsJson = (syncData['products'] as List? ?? []).cast<Map<String, dynamic>>();
        final deletedIds = (syncData['deleted_ids'] as List? ?? []).cast<String>();
        final newMaxSeq = (syncData['new_max_seq'] as num).toInt();
        hasMore = syncData['has_more'] as bool? ?? false;

        await database.batchWriteSyncPage(
          storeId: storeId,
          productsJson: productsJson,
          deletedIds: deletedIds,
          newMaxSeq: newMaxSeq,
        );

        pagesSynced++;
        sinceSeq = newMaxSeq;

        _progressController.add(SyncProgress(
          pagesSynced: pagesSynced,
          done: !hasMore,
          currentMaxSeq: newMaxSeq,
        ));

        if (kDebugMode) {
          print('[CATALOG_SYNC] Batch page $pagesSynced committed (products: ${productsJson.length}, deleted: ${deletedIds.length}, maxSeq: $newMaxSeq, hasMore: $hasMore)');
        }
      } catch (e) {
        if (kDebugMode) {
          print('[CATALOG_SYNC] Sync interrupted on page $pagesSynced: $e');
        }
        rethrow;
      }
    }

    if (kDebugMode) {
      print('[CATALOG_SYNC] Delta sync complete for store $storeId! Total pages: $pagesSynced');
    }
  }
}

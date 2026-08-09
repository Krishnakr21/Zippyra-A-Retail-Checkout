import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/catalog/data/catalog_sync_engine.dart';
import 'package:customer_app/features/catalog/data/datasources/catalog_remote_data_source.dart';
import 'package:customer_app/features/catalog/data/drift/catalog_database.dart';
import 'package:customer_app/features/catalog/domain/entities/category.dart';
import 'package:customer_app/features/catalog/domain/entities/product.dart';

class FakeRemoteDataSource implements CatalogRemoteDataSource {
  final List<int> requestedSinceSeqs = [];

  @override
  Future<Map<String, dynamic>> postDeltaSync(String storeId, int sinceSeq) async {
    requestedSinceSeqs.add(sinceSeq);

    if (sinceSeq == 0) {
      // Page 1
      return {
        'products': [
          {
            'id': 'p1',
            'store_id': storeId,
            'chain_id': 'chain-1',
            'barcode': '8901030300011',
            'name': 'Item 1',
            'price_paise': 100,
            'mrp_paise': 100,
            'hsn_code': '0901',
            'sync_seq': 10,
          }
        ],
        'deleted_ids': <String>[],
        'new_max_seq': 10,
        'has_more': true,
      };
    } else if (sinceSeq == 10) {
      // Page 2
      return {
        'products': [
          {
            'id': 'p2',
            'store_id': storeId,
            'chain_id': 'chain-1',
            'barcode': '012345678905',
            'name': 'Item 2',
            'price_paise': 200,
            'mrp_paise': 200,
            'hsn_code': '1905',
            'sync_seq': 20,
          }
        ],
        'deleted_ids': <String>[],
        'new_max_seq': 20,
        'has_more': false,
      };
    }

    return {
      'products': <Map<String, dynamic>>[],
      'deleted_ids': <String>[],
      'new_max_seq': sinceSeq,
      'has_more': false,
    };
  }

  @override
  Future<Product> getProductByBarcode(String storeId, String barcode) async => throw UnimplementedError();

  @override
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1}) async => [];

  @override
  Future<List<Category>> getCategories(String chainId) async => [];
}

class FakeDatabase implements CatalogDatabase {
  int lastSyncSeq = 0;

  @override
  Future<int> getLastSyncSeq(String storeId) async => lastSyncSeq;

  @override
  Future<void> batchWriteSyncPage({
    required String storeId,
    required List<Map<String, dynamic>> productsJson,
    required List<String> deletedIds,
    required int newMaxSeq,
  }) async {
    lastSyncSeq = newMaxSeq;
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

void main() {
  late FakeRemoteDataSource fakeRemote;
  late FakeDatabase fakeDb;
  late CatalogSyncEngine syncEngine;

  setUp(() {
    fakeRemote = FakeRemoteDataSource();
    fakeDb = FakeDatabase();
    syncEngine = CatalogSyncEngine(remoteDataSource: fakeRemote, database: fakeDb);
  });

  test('resumes delta sync from lastSyncSeq after crash/restart', () async {
    // 1. Simulate page 1 completed before a crash
    fakeDb.lastSyncSeq = 10;

    // 2. Restart sync
    await syncEngine.syncCatalog('store-1');

    // Assert that the sync engine resumed from since_seq = 10 (not 0!)
    expect(fakeRemote.requestedSinceSeqs.first, equals(10));
    expect(fakeDb.lastSyncSeq, equals(20));
  });
}

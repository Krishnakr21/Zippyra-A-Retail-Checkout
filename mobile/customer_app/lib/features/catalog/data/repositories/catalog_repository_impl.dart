import 'package:flutter/foundation.dart' hide Category;
import '../../domain/entities/category.dart';
import '../../domain/entities/product.dart';
import '../../domain/repositories/catalog_repository.dart';
import '../catalog_sync_engine.dart';
import '../datasources/catalog_local_data_source.dart';
import '../datasources/catalog_remote_data_source.dart';

class CatalogRepositoryImpl implements CatalogRepository {
  final CatalogLocalDataSource localDataSource;
  final CatalogRemoteDataSource remoteDataSource;
  final CatalogSyncEngine syncEngine;

  CatalogRepositoryImpl({
    required this.localDataSource,
    required this.remoteDataSource,
    required this.syncEngine,
  });

  @override
  Future<Product?> getProductByBarcode(String storeId, String barcode) async {
    // 1. LOCAL-FIRST Query
    final localProduct = await localDataSource.getProductByBarcode(storeId, barcode);
    if (localProduct != null) {
      return localProduct;
    }

    // Telemetry log on local cache miss fallback to remote
    if (kDebugMode) {
      print('[TELEMETRY] Local catalog cache miss for barcode $barcode at store $storeId. Falling back to remote API...');
    }

    // 2. REMOTE Fallback
    try {
      return await remoteDataSource.getProductByBarcode(storeId, barcode);
    } catch (_) {
      return null;
    }
  }

  @override
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1}) async {
    // LOCAL-FIRST Query against Drift database
    final localProducts = await localDataSource.searchProducts(storeId, query, categoryId: categoryId, page: page);
    if (localProducts.isNotEmpty) {
      return localProducts;
    }

    // Remote query fallback if local results empty
    try {
      return await remoteDataSource.searchProducts(storeId, query, categoryId: categoryId, page: page);
    } catch (_) {
      return [];
    }
  }

  @override
  Future<List<Category>> getCategories(String chainId) async {
    final localCategories = await localDataSource.getCategories(chainId);
    if (localCategories.isNotEmpty) {
      return localCategories;
    }

    try {
      return await remoteDataSource.getCategories(chainId);
    } catch (_) {
      return [];
    }
  }

  @override
  Future<void> syncCatalog(String storeId, {bool force = false}) {
    return syncEngine.syncCatalog(storeId, force: force);
  }
}

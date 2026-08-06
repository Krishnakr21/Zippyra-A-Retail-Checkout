import '../drift/catalog_database.dart';
import '../../domain/entities/category.dart';
import '../../domain/entities/product.dart';

abstract class CatalogLocalDataSource {
  Future<Product?> getProductByBarcode(String storeId, String barcode);
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1});
  Future<List<Category>> getCategories(String chainId);
  Future<int> getLastSyncSeq(String storeId);
}

class CatalogLocalDataSourceImpl implements CatalogLocalDataSource {
  final CatalogDatabase database;

  CatalogLocalDataSourceImpl({CatalogDatabase? database})
      : database = database ?? CatalogDatabase.instance;

  @override
  Future<Product?> getProductByBarcode(String storeId, String barcode) {
    return database.getProductByBarcode(storeId, barcode);
  }

  @override
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1}) {
    return database.searchProducts(storeId, query, categoryId: categoryId, page: page);
  }

  @override
  Future<List<Category>> getCategories(String chainId) {
    return database.getCategories(chainId);
  }

  @override
  Future<int> getLastSyncSeq(String storeId) {
    return database.getLastSyncSeq(storeId);
  }
}

import '../entities/category.dart';
import '../entities/product.dart';

abstract class CatalogRepository {
  Future<Product?> getProductByBarcode(String storeId, String barcode);
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1});
  Future<List<Category>> getCategories(String chainId);
  Future<void> syncCatalog(String storeId, {bool force = false});
}

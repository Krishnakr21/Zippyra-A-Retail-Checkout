import '../entities/product.dart';
import '../repositories/catalog_repository.dart';

class SearchProductsUseCase {
  final CatalogRepository repository;

  const SearchProductsUseCase(this.repository);

  Future<List<Product>> call(String storeId, String query, {String? categoryId, int page = 1}) {
    return repository.searchProducts(storeId, query, categoryId: categoryId, page: page);
  }
}

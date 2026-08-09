import '../entities/category.dart';
import '../repositories/catalog_repository.dart';

class GetCategoriesUseCase {
  final CatalogRepository repository;

  const GetCategoriesUseCase(this.repository);

  Future<List<Category>> call(String chainId) {
    return repository.getCategories(chainId);
  }
}

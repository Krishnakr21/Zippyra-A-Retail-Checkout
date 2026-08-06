import '../repositories/catalog_repository.dart';

class SyncCatalogUseCase {
  final CatalogRepository repository;

  const SyncCatalogUseCase(this.repository);

  Future<void> call(String storeId, {bool force = false}) {
    return repository.syncCatalog(storeId, force: force);
  }
}

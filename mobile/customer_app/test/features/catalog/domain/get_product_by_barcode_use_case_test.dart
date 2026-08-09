import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/catalog/domain/entities/category.dart';
import 'package:customer_app/features/catalog/domain/entities/product.dart';
import 'package:customer_app/features/catalog/domain/repositories/catalog_repository.dart';
import 'package:customer_app/features/catalog/domain/usecases/get_product_by_barcode_use_case.dart';

class MockCatalogRepository implements CatalogRepository {
  bool localHit = true;
  int remoteCallCount = 0;

  @override
  Future<Product?> getProductByBarcode(String storeId, String barcode) async {
    if (localHit) {
      return const Product(
        id: 'p1',
        barcode: '8901030300011',
        name: 'Local Coffee',
        pricePaise: 25000,
        mrpPaise: 28000,
      );
    }
    remoteCallCount++;
    return null;
  }

  @override
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1}) async => [];

  @override
  Future<List<Category>> getCategories(String chainId) async => [];

  @override
  Future<void> syncCatalog(String storeId, {bool force = false}) async {}
}

void main() {
  late MockCatalogRepository mockRepo;
  late GetProductByBarcodeUseCase useCase;

  setUp(() {
    mockRepo = MockCatalogRepository();
    useCase = GetProductByBarcodeUseCase(mockRepo);
  });

  test('returns product from local cache without triggering remote call when row exists', () async {
    mockRepo.localHit = true;
    final product = await useCase('store-1', '8901030300011');

    expect(product, isNotNull);
    expect(product!.name, 'Local Coffee');
    expect(mockRepo.remoteCallCount, equals(0));
  });
}

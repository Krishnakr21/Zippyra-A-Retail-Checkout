import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/catalog/domain/entities/category.dart';
import 'package:customer_app/features/catalog/domain/entities/product.dart';
import 'package:customer_app/features/catalog/domain/repositories/catalog_repository.dart';
import 'package:customer_app/features/catalog/domain/usecases/get_categories_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/get_product_by_barcode_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/search_products_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/sync_catalog_use_case.dart';
import 'package:customer_app/features/catalog/presentation/bloc/catalog_bloc.dart';

class FakeSearchProductsUseCase implements SearchProductsUseCase {
  int callCount = 0;
  List<Product> returnProducts = [];

  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<List<Product>> call(String storeId, String query, {String? categoryId, int page = 1}) async {
    callCount++;
    return returnProducts;
  }
}

class FakeGetCategoriesUseCase implements GetCategoriesUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<List<Category>> call(String chainId) async => [];
}

class FakeGetProductByBarcodeUseCase implements GetProductByBarcodeUseCase {
  Product? mockProduct;

  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<Product?> call(String storeId, String barcode) async => mockProduct;
}

class FakeSyncCatalogUseCase implements SyncCatalogUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<void> call(String storeId, {bool force = false}) async {}
}

void main() {
  late FakeSearchProductsUseCase fakeSearch;
  late FakeGetProductByBarcodeUseCase fakeBarcode;
  late CatalogBloc bloc;

  setUp(() {
    fakeSearch = FakeSearchProductsUseCase();
    fakeBarcode = FakeGetProductByBarcodeUseCase();
    bloc = CatalogBloc(
      searchProductsUseCase: fakeSearch,
      getCategoriesUseCase: FakeGetCategoriesUseCase(),
      getProductByBarcodeUseCase: fakeBarcode,
      syncCatalogUseCase: FakeSyncCatalogUseCase(),
    );
  });

  tearDown(() {
    bloc.close();
  });

  blocTest<CatalogBloc, CatalogState>(
    'SearchQueryChanged emits [CatalogSearching, CatalogEmpty] when search results empty',
    build: () => bloc,
    act: (bloc) => bloc.add(const SearchQueryChanged(query: 'Coffee', storeId: 'store-1')),
    wait: const Duration(milliseconds: 350),
    expect: () => [
      isA<CatalogSearching>(),
      isA<CatalogEmpty>(),
    ],
  );

  blocTest<CatalogBloc, CatalogState>(
    'ProductRequested emits [CatalogSearching, CatalogProductFound] when barcode exists',
    build: () {
      fakeBarcode.mockProduct = const Product(
        id: 'p1',
        barcode: '8901030300011',
        name: 'Found Coffee',
        pricePaise: 2000,
        mrpPaise: 2000,
      );
      return bloc;
    },
    act: (bloc) => bloc.add(const ProductRequested(storeId: 'store-1', barcode: '8901030300011')),
    expect: () => [
      isA<CatalogSearching>(),
      isA<CatalogProductFound>(),
    ],
  );
}

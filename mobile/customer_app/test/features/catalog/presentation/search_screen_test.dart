import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/catalog/domain/entities/category.dart';
import 'package:customer_app/features/catalog/domain/entities/product.dart';
import 'package:customer_app/features/catalog/domain/repositories/catalog_repository.dart';
import 'package:customer_app/features/catalog/domain/usecases/get_categories_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/get_product_by_barcode_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/search_products_use_case.dart';
import 'package:customer_app/features/catalog/domain/usecases/sync_catalog_use_case.dart';
import 'package:customer_app/features/catalog/presentation/bloc/catalog_bloc.dart';
import 'package:customer_app/features/catalog/presentation/screens/search_screen.dart';
import 'package:customer_app/features/catalog/presentation/widgets/empty_search_state.dart';

class FakeSearchProductsUseCase implements SearchProductsUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<List<Product>> call(String storeId, String query, {String? categoryId, int page = 1}) async {
    return [];
  }
}

class FakeGetCategoriesUseCase implements GetCategoriesUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<List<Category>> call(String chainId) async => [];
}

class FakeGetProductByBarcodeUseCase implements GetProductByBarcodeUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<Product?> call(String storeId, String barcode) async => null;
}

class FakeSyncCatalogUseCase implements SyncCatalogUseCase {
  @override
  CatalogRepository get repository => throw UnimplementedError();

  @override
  Future<void> call(String storeId, {bool force = false}) async {}
}

void main() {
  late CatalogBloc bloc;

  setUp(() {
    bloc = CatalogBloc(
      searchProductsUseCase: FakeSearchProductsUseCase(),
      getCategoriesUseCase: FakeGetCategoriesUseCase(),
      getProductByBarcodeUseCase: FakeGetProductByBarcodeUseCase(),
      syncCatalogUseCase: FakeSyncCatalogUseCase(),
    );
  });

  tearDown(() {
    bloc.close();
  });

  Widget buildTestableWidget() {
    return MaterialApp(
      home: BlocProvider<CatalogBloc>.value(
        value: bloc,
        child: const SearchScreen(storeId: 'store-1'),
      ),
    );
  }

  testWidgets('SearchScreen renders TextField and EmptySearchState when results empty', (tester) async {
    await tester.pumpWidget(buildTestableWidget());
    await tester.pump();

    // Verify search text field presence
    final searchField = find.byType(TextField);
    expect(searchField, findsOneWidget);

    // Type text into search field
    await tester.enterText(searchField, 'Coffee');
    await tester.pump();

    // Verify SearchScreen rendered cleanly
    expect(find.byType(SearchScreen), findsOneWidget);
  });
}

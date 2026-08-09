import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:stream_transform/stream_transform.dart';
import '../../domain/entities/category.dart';
import '../../domain/entities/product.dart';
import '../../domain/usecases/get_categories_use_case.dart';
import '../../domain/usecases/get_product_by_barcode_use_case.dart';
import '../../domain/usecases/search_products_use_case.dart';
import '../../domain/usecases/sync_catalog_use_case.dart';

// --- Events ---
abstract class CatalogEvent {
  const CatalogEvent();
}

class SearchQueryChanged extends CatalogEvent {
  final String query;
  final String storeId;
  final String? categoryId;

  const SearchQueryChanged({
    required this.query,
    required this.storeId,
    this.categoryId,
  });
}

class CategorySelected extends CatalogEvent {
  final String? categoryId;
  final String storeId;
  final String query;

  const CategorySelected({
    required this.categoryId,
    required this.storeId,
    required this.query,
  });
}

class ProductRequested extends CatalogEvent {
  final String storeId;
  final String barcode;

  const ProductRequested({required this.storeId, required this.barcode});
}

class CatalogSyncRequested extends CatalogEvent {
  final String storeId;
  final bool force;

  const CatalogSyncRequested({required this.storeId, this.force = false});
}

// --- States ---
abstract class CatalogState {
  const CatalogState();
}

class CatalogIdle extends CatalogState {}

class CatalogSearching extends CatalogState {}

class CatalogResults extends CatalogState {
  final List<Product> products;
  final List<Category> categories;
  final String? selectedCategoryId;

  const CatalogResults({
    required this.products,
    this.categories = const [],
    this.selectedCategoryId,
  });
}

class CatalogEmpty extends CatalogState {
  final String query;
  final List<Category> categories;
  final String? selectedCategoryId;

  const CatalogEmpty({
    required this.query,
    this.categories = const [],
    this.selectedCategoryId,
  });
}

class CatalogProductFound extends CatalogState {
  final Product product;
  const CatalogProductFound(this.product);
}

class CatalogProductNotFound extends CatalogState {
  final String barcode;
  const CatalogProductNotFound(this.barcode);
}

// --- BLoC ---
EventTransformer<Event> debounce<Event>(Duration duration) {
  return (events, mapper) => events.debounce(duration).switchMap(mapper);
}

class CatalogBloc extends Bloc<CatalogEvent, CatalogState> {
  final GetProductByBarcodeUseCase getProductByBarcodeUseCase;
  final SearchProductsUseCase searchProductsUseCase;
  final GetCategoriesUseCase getCategoriesUseCase;
  final SyncCatalogUseCase syncCatalogUseCase;

  List<Category> _cachedCategories = [];

  CatalogBloc({
    required this.getProductByBarcodeUseCase,
    required this.searchProductsUseCase,
    required this.getCategoriesUseCase,
    required this.syncCatalogUseCase,
  }) : super(CatalogIdle()) {
    on<SearchQueryChanged>(_onSearchQueryChanged, transformer: debounce(const Duration(milliseconds: 300)));
    on<CategorySelected>(_onCategorySelected);
    on<ProductRequested>(_onProductRequested);
    on<CatalogSyncRequested>(_onCatalogSyncRequested);
  }

  Future<void> _onSearchQueryChanged(
    SearchQueryChanged event,
    Emitter<CatalogState> emit,
  ) async {
    emit(CatalogSearching());
    await _performSearch(event.storeId, event.query, event.categoryId, emit);
  }

  Future<void> _onCategorySelected(
    CategorySelected event,
    Emitter<CatalogState> emit,
  ) async {
    emit(CatalogSearching());
    await _performSearch(event.storeId, event.query, event.categoryId, emit);
  }

  Future<void> _performSearch(
    String storeId,
    String query,
    String? categoryId,
    Emitter<CatalogState> emit,
  ) async {
    if (_cachedCategories.isEmpty) {
      _cachedCategories = await getCategoriesUseCase('chain-hq-001');
    }

    final products = await searchProductsUseCase(storeId, query, categoryId: categoryId);
    if (products.isEmpty) {
      emit(CatalogEmpty(
        query: query,
        categories: _cachedCategories,
        selectedCategoryId: categoryId,
      ));
    } else {
      emit(CatalogResults(
        products: products,
        categories: _cachedCategories,
        selectedCategoryId: categoryId,
      ));
    }
  }

  Future<void> _onProductRequested(
    ProductRequested event,
    Emitter<CatalogState> emit,
  ) async {
    emit(CatalogSearching());
    final product = await getProductByBarcodeUseCase(event.storeId, event.barcode);
    if (product != null) {
      emit(CatalogProductFound(product));
    } else {
      emit(CatalogProductNotFound(event.barcode));
    }
  }

  Future<void> _onCatalogSyncRequested(
    CatalogSyncRequested event,
    Emitter<CatalogState> emit,
  ) async {
    await syncCatalogUseCase(event.storeId, force: event.force);
  }
}

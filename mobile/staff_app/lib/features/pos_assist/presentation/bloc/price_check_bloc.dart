import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'price_check_event.dart';
import 'price_check_state.dart';

class PriceCheckBloc extends Bloc<PriceCheckEvent, PriceCheckState> {
  final CatalogDatabase catalogDatabase;
  final ApiClient apiClient;

  PriceCheckBloc({
    CatalogDatabase? catalogDatabase,
    required this.apiClient,
  })  : catalogDatabase = catalogDatabase ?? CatalogDatabase.instance,
        super(PriceCheckInitial()) {
    on<BarcodeScanned>(_onBarcodeScanned);
    on<ManualBarcodeSubmitted>(_onBarcodeScanned);
    on<PriceCheckReset>((event, emit) => emit(PriceCheckInitial()));
  }

  Future<void> _onBarcodeScanned(
    PriceCheckEvent event,
    Emitter<PriceCheckState> emit,
  ) async {
    final storeId = event is BarcodeScanned ? event.storeId : (event as ManualBarcodeSubmitted).storeId;
    final barcode = event is BarcodeScanned ? event.barcode.trim() : (event as ManualBarcodeSubmitted).barcode.trim();

    if (barcode.isEmpty) return;

    emit(PriceCheckLoading());

    try {
      // 1. LOCAL FIRST LOOKUP
      final localProduct = await catalogDatabase.getProductByBarcode(storeId, barcode);
      if (localProduct != null) {
        emit(PriceCheckFound(product: localProduct, fetchedFromRemote: false));
        return;
      }

      // 2. REMOTE FALLBACK LOOKUP
      final response = await apiClient.get('/v1/catalog/barcode/$barcode', queryParameters: {'store_id': storeId});
      if (response.statusCode == 200 && response.data != null) {
        final data = response.data as Map<String, dynamic>;
        final remoteProduct = SharedCatalogProduct(
          id: data['id'] as String,
          barcode: data['barcode'] as String,
          name: data['name'] as String,
          description: data['description'] as String? ?? '',
          pricePaise: (data['price_paise'] as num).toInt(),
          mrpPaise: (data['mrp_paise'] as num).toInt(),
          hsnCode: data['hsn_code'] as String? ?? '',
          gstRatePercent: (data['gst_rate_percent'] as num).toDouble(),
          imageUrl: data['image_url'] as String? ?? '',
          thumbnailUrl: data['thumbnail_url'] as String? ?? '',
          isReturnable: data['is_returnable'] as bool? ?? true,
        );
        emit(PriceCheckFound(product: remoteProduct, fetchedFromRemote: true));
        return;
      }

      emit(PriceCheckNotFound(barcode));
    } catch (e) {
      if ((e is DioException && e.response?.statusCode == 404) ||
          e.toString().contains('404') ||
          e.toString().contains('PRODUCT_NOT_FOUND')) {
        emit(PriceCheckNotFound(barcode));
      } else {
        emit(PriceCheckFailed('Error checking product: ${e.toString()}'));
      }
    }
  }
}

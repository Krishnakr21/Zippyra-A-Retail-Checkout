import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:staff_app/features/pos_assist/presentation/bloc/price_check_bloc.dart';
import 'package:staff_app/features/pos_assist/presentation/bloc/price_check_event.dart';
import 'package:staff_app/features/pos_assist/presentation/bloc/price_check_state.dart';

class FakeApiClient extends Fake implements ApiClient {
  int remoteCalls = 0;

  @override
  Future<Response> get(String path, {Map<String, dynamic>? queryParameters}) async {
    remoteCalls++;
    if (path.contains('8901000000002')) {
      return Response(
        requestOptions: RequestOptions(path: path),
        statusCode: 200,
        data: {
          'id': 'p-remote-2',
          'barcode': '8901000000002',
          'name': 'Remote Organic Milk',
          'description': '',
          'price_paise': 6500,
          'mrp_paise': 7000,
          'hsn_code': '0401',
          'gst_rate_percent': 5.0,
          'image_url': '',
          'thumbnail_url': '',
          'is_returnable': true,
        },
      );
    }
    throw DioException(
      requestOptions: RequestOptions(path: path),
      response: Response(
        requestOptions: RequestOptions(path: path),
        statusCode: 404,
        data: {'error': {'code': 'PRODUCT_NOT_FOUND', 'message': 'Product not found'}},
      ),
    );
  }
}

class FakeCatalogDatabase extends Fake implements CatalogDatabase {
  final Map<String, SharedCatalogProduct> localStore = {};

  @override
  Future<SharedCatalogProduct?> getProductByBarcode(String storeId, String barcode) async {
    return localStore['$storeId:$barcode'];
  }
}

void main() {
  late FakeApiClient fakeApiClient;
  late FakeCatalogDatabase fakeDb;
  late PriceCheckBloc bloc;

  setUp(() {
    fakeApiClient = FakeApiClient();
    fakeDb = FakeCatalogDatabase();
    bloc = PriceCheckBloc(
      catalogDatabase: fakeDb,
      apiClient: fakeApiClient,
    );
  });

  group('PriceCheckBloc Tests', () {
    test('Barcode present in local cache resolves without remote call', () async {
      fakeDb.localStore['store-1:8901000000001'] = SharedCatalogProduct(
        id: 'p-local-1',
        barcode: '8901000000001',
        name: 'Local Fresh Bread',
        description: '',
        pricePaise: 4000,
        mrpPaise: 4500,
        hsnCode: '1905',
        gstRatePercent: 0.0,
        imageUrl: '',
        thumbnailUrl: '',
        isReturnable: false,
      );

      bloc.add(const BarcodeScanned(storeId: 'store-1', barcode: '8901000000001'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          PriceCheckLoading(),
          predicate<PriceCheckFound>((s) {
            return s.product.name == 'Local Fresh Bread' && s.fetchedFromRemote == false;
          }),
        ]),
      );

      expect(fakeApiClient.remoteCalls, 0); // Local-first: 0 remote calls
    });

    test('Barcode absent locally falls back to remote endpoint', () async {
      bloc.add(const BarcodeScanned(storeId: 'store-1', barcode: '8901000000002'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          PriceCheckLoading(),
          predicate<PriceCheckFound>((s) {
            return s.product.name == 'Remote Organic Milk' && s.fetchedFromRemote == true;
          }),
        ]),
      );

      expect(fakeApiClient.remoteCalls, 1); // Remote fallback executed once
    });

    test('Unknown barcode returns PriceCheckNotFound state', () async {
      bloc.add(const BarcodeScanned(storeId: 'store-1', barcode: '9999999999999'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          PriceCheckLoading(),
          predicate<PriceCheckNotFound>((s) => s.barcode == '9999999999999'),
        ]),
      );

      expect(fakeApiClient.remoteCalls, 1);
    });
  });
}

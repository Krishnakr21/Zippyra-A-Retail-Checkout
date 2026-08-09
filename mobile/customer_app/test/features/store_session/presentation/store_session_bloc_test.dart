import 'package:flutter_test/flutter_test.dart';
import 'package:bloc_test/bloc_test.dart';
import 'package:customer_app/features/store_session/domain/entities/nearby_store.dart';
import 'package:customer_app/features/store_session/domain/entities/store_session.dart';
import 'package:customer_app/features/store_session/domain/repositories/store_session_repository.dart';
import 'package:customer_app/features/store_session/domain/usecases/bind_store_use_case.dart';
import 'package:customer_app/features/store_session/domain/usecases/get_nearby_stores_use_case.dart';
import 'package:customer_app/features/store_session/domain/usecases/restore_session_use_case.dart';
import 'package:customer_app/features/store_session/domain/usecases/unbind_store_use_case.dart';
import 'package:customer_app/features/store_session/presentation/bloc/store_session_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';

class MockStoreSessionRepository implements StoreSessionRepository {
  bool shouldRestore = false;
  Failure? bindFailure;
  StoreSession? activeSession;

  @override
  Future<List<NearbyStore>> getNearbyStores(double lat, double lng, {double radiusKm = 10.0}) async {
    return const [
      NearbyStore(id: 's1', name: 'Store 1', address: 'Bangalore', distanceKm: 1.2, isOpen: true, capacityPct: 45),
    ];
  }

  @override
  Future<StoreSession> bindStore({
    required String qrToken,
    required double lat,
    required double lng,
    required String deviceId,
  }) async {
    if (bindFailure != null) {
      throw bindFailure!;
    }
    return activeSession ??
        const StoreSession(
          storeId: 's1',
          storeName: 'Store 1',
          sessionToken: 'sess_token_123',
          catalogVersion: 5,
          expiresAt: '2026-07-31T06:00:00Z',
        );
  }

  @override
  Future<void> unbindStore({String? sessionId, String reason = 'customer_exit'}) async {
    activeSession = null;
  }

  @override
  Future<StoreSession?> restoreSession() async {
    if (shouldRestore) {
      return activeSession ??
          const StoreSession(
            storeId: 's1',
            storeName: 'Store 1',
            sessionToken: 'sess_token_123',
            catalogVersion: 5,
            expiresAt: '2026-07-31T06:00:00Z',
          );
    }
    return null;
  }
}

class MockSecureStorage extends SecureStorage {
  @override
  Future<String> getDeviceId() async => 'test-device-id';
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockStoreSessionRepository mockRepo;
  late MockSecureStorage mockStorage;
  late GetNearbyStoresUseCase getNearbyStoresUseCase;
  late BindStoreUseCase bindStoreUseCase;
  late UnbindStoreUseCase unbindStoreUseCase;
  late RestoreSessionUseCase restoreSessionUseCase;
  late StoreSessionBloc bloc;

  setUp(() {
    mockRepo = MockStoreSessionRepository();
    mockStorage = MockSecureStorage();
    getNearbyStoresUseCase = GetNearbyStoresUseCase(mockRepo);
    bindStoreUseCase = BindStoreUseCase(mockRepo);
    unbindStoreUseCase = UnbindStoreUseCase(mockRepo);
    restoreSessionUseCase = RestoreSessionUseCase(mockRepo);

    bloc = StoreSessionBloc(
      getNearbyStoresUseCase: getNearbyStoresUseCase,
      bindStoreUseCase: bindStoreUseCase,
      unbindStoreUseCase: unbindStoreUseCase,
      restoreSessionUseCase: restoreSessionUseCase,
      secureStorage: mockStorage,
    );
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is StoreSessionInitial', () {
    expect(bloc.state, isA<StoreSessionInitial>());
  });

  blocTest<StoreSessionBloc, StoreSessionState>(
    'AppStartedSessionCheckRequested skips scan flow when session exists',
    build: () {
      mockRepo.shouldRestore = true;
      return bloc;
    },
    act: (bloc) => bloc.add(AppStartedSessionCheckRequested()),
    expect: () => [
      isA<StoreSessionRestoring>(),
      isA<StoreSessionActive>().having((s) => s.session.catalogVersion, 'catalogVersion', 5),
    ],
  );

  blocTest<StoreSessionBloc, StoreSessionState>(
    'BindRequested success emits StoreSessionBinding then StoreSessionActive',
    build: () => bloc,
    act: (bloc) => bloc.add(const BindRequested(qrToken: 'valid_qr_token', lat: 12.9716, lng: 77.5946)),
    expect: () => [
      isA<StoreSessionBinding>(),
      isA<StoreSessionActive>().having((s) => s.session.storeId, 'storeId', 's1'),
    ],
  );

  blocTest<StoreSessionBloc, StoreSessionState>(
    'BindRequested STORE_CLOSED error maps to StoreClosedFailure',
    build: () {
      mockRepo.bindFailure = const StoreClosedFailure('Store closed');
      return bloc;
    },
    act: (bloc) => bloc.add(const BindRequested(qrToken: 'closed_qr_token', lat: 12.9716, lng: 77.5946)),
    expect: () => [
      isA<StoreSessionBinding>(),
      isA<StoreSessionBindFailure>().having((s) => s.failure, 'failure', isA<StoreClosedFailure>()),
    ],
  );

  blocTest<StoreSessionBloc, StoreSessionState>(
    'BindRequested STORE_AT_CAPACITY error maps to StoreAtCapacityFailure',
    build: () {
      mockRepo.bindFailure = const StoreAtCapacityFailure('Store full');
      return bloc;
    },
    act: (bloc) => bloc.add(const BindRequested(qrToken: 'full_qr_token', lat: 12.9716, lng: 77.5946)),
    expect: () => [
      isA<StoreSessionBinding>(),
      isA<StoreSessionBindFailure>().having((s) => s.failure, 'failure', isA<StoreAtCapacityFailure>()),
    ],
  );

  blocTest<StoreSessionBloc, StoreSessionState>(
    'BindRequested STORE_GEOFENCE_MISMATCH maps to StoreGeofenceMismatchFailure',
    build: () {
      mockRepo.bindFailure = const StoreGeofenceMismatchFailure('Outside geofence');
      return bloc;
    },
    act: (bloc) => bloc.add(const BindRequested(qrToken: 'geofence_qr_token', lat: 12.9716, lng: 77.5946)),
    expect: () => [
      isA<StoreSessionBinding>(),
      isA<StoreSessionBindFailure>().having((s) => s.failure, 'failure', isA<StoreGeofenceMismatchFailure>()),
    ],
  );
}

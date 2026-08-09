import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:geolocator/geolocator.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../../core/services/sync_engine.dart';
import '../../domain/entities/nearby_store.dart';
import '../../domain/entities/store_session.dart';
import '../../domain/usecases/bind_store_use_case.dart';
import '../../domain/usecases/get_nearby_stores_use_case.dart';
import '../../domain/usecases/restore_session_use_case.dart';
import '../../domain/usecases/unbind_store_use_case.dart';

// --- Events ---
abstract class StoreSessionEvent {
  const StoreSessionEvent();
}

class AppStartedSessionCheckRequested extends StoreSessionEvent {}

class NearbyStoresRequested extends StoreSessionEvent {
  final double lat;
  final double lng;
  final double radiusKm;

  const NearbyStoresRequested({required this.lat, required this.lng, this.radiusKm = 10.0});
}

class EntranceQrScanned extends StoreSessionEvent {
  final String qrToken;
  const EntranceQrScanned(this.qrToken);
}

class BindRequested extends StoreSessionEvent {
  final String? qrToken;
  final double? lat;
  final double? lng;

  const BindRequested({this.qrToken, this.lat, this.lng});
}

class UnbindRequested extends StoreSessionEvent {
  final String reason;
  const UnbindRequested({this.reason = 'customer_exit'});
}

class CapacityRetryTimerFired extends StoreSessionEvent {}

// --- States ---
abstract class StoreSessionState {
  const StoreSessionState();
}

class StoreSessionInitial extends StoreSessionState {}

class StoreSessionRestoring extends StoreSessionState {}

class StoreSessionActive extends StoreSessionState {
  final StoreSession session;
  final List<NearbyStore> nearbyStores;

  const StoreSessionActive({required this.session, this.nearbyStores = const []});
}

class StoreSessionNone extends StoreSessionState {
  final List<NearbyStore> nearbyStores;
  final bool isLoadingStores;

  const StoreSessionNone({this.nearbyStores = const [], this.isLoadingStores = false});
}

class StoreSessionBinding extends StoreSessionState {
  final String? qrToken;
  const StoreSessionBinding({this.qrToken});
}

class StoreSessionBindFailure extends StoreSessionState {
  final Failure failure;
  final String? qrToken;

  const StoreSessionBindFailure({required this.failure, this.qrToken});
}

class StoreSessionUnbinding extends StoreSessionState {}

// --- BLoC ---
class StoreSessionBloc extends Bloc<StoreSessionEvent, StoreSessionState> {
  final GetNearbyStoresUseCase getNearbyStoresUseCase;
  final BindStoreUseCase bindStoreUseCase;
  final UnbindStoreUseCase unbindStoreUseCase;
  final RestoreSessionUseCase restoreSessionUseCase;
  final SecureStorage secureStorage;

  String? _lastScannedQrToken;

  StoreSessionBloc({
    required this.getNearbyStoresUseCase,
    required this.bindStoreUseCase,
    required this.unbindStoreUseCase,
    required this.restoreSessionUseCase,
    required this.secureStorage,
  }) : super(StoreSessionInitial()) {
    on<AppStartedSessionCheckRequested>(_onAppStartedSessionCheckRequested);
    on<NearbyStoresRequested>(_onNearbyStoresRequested);
    on<EntranceQrScanned>(_onEntranceQrScanned);
    on<BindRequested>(_onBindRequested);
    on<UnbindRequested>(_onUnbindRequested);
    on<CapacityRetryTimerFired>(_onCapacityRetryTimerFired);
  }

  Future<void> _onAppStartedSessionCheckRequested(
    AppStartedSessionCheckRequested event,
    Emitter<StoreSessionState> emit,
  ) async {
    emit(StoreSessionRestoring());
    try {
      final session = await restoreSessionUseCase();
      if (session != null) {
        // Trigger background catalog sync without blocking
        SyncEngine().triggerCatalogSync(session.catalogVersion);
        emit(StoreSessionActive(session: session));
      } else {
        emit(const StoreSessionNone());
      }
    } catch (_) {
      emit(const StoreSessionNone());
    }
  }

  Future<void> _onNearbyStoresRequested(
    NearbyStoresRequested event,
    Emitter<StoreSessionState> emit,
  ) async {
    final currentSession = state is StoreSessionActive ? (state as StoreSessionActive).session : null;
    if (state is StoreSessionNone) {
      emit(StoreSessionNone(
        nearbyStores: (state as StoreSessionNone).nearbyStores,
        isLoadingStores: true,
      ));
    }

    try {
      final stores = await getNearbyStoresUseCase(event.lat, event.lng, radiusKm: event.radiusKm);
      if (currentSession != null) {
        emit(StoreSessionActive(session: currentSession, nearbyStores: stores));
      } else {
        emit(StoreSessionNone(nearbyStores: stores, isLoadingStores: false));
      }
    } catch (_) {
      if (currentSession != null) {
        emit(StoreSessionActive(session: currentSession));
      } else {
        emit(const StoreSessionNone(isLoadingStores: false));
      }
    }
  }

  void _onEntranceQrScanned(
    EntranceQrScanned event,
    Emitter<StoreSessionState> emit,
  ) {
    _lastScannedQrToken = event.qrToken;
    add(BindRequested(qrToken: event.qrToken));
  }

  Future<void> _onBindRequested(
    BindRequested event,
    Emitter<StoreSessionState> emit,
  ) async {
    final qrToken = event.qrToken ?? _lastScannedQrToken;
    if (qrToken == null || qrToken.isEmpty) {
      emit(const StoreSessionBindFailure(failure: QRTokenInvalidFailure()));
      return;
    }

    emit(StoreSessionBinding(qrToken: qrToken));

    // Get GPS coordinates with timeout and fallback
    double lat = event.lat ?? 12.9716;
    double lng = event.lng ?? 77.5946;

    if (event.lat == null || event.lng == null) {
      try {
        final pos = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.best,
          timeLimit: const Duration(seconds: 10),
        );
        lat = pos.latitude;
        lng = pos.longitude;
      } catch (_) {
        // Fallback to last known position if current fix fails/times out
        try {
          final lastPos = await Geolocator.getLastKnownPosition();
          if (lastPos != null) {
            lat = lastPos.latitude;
            lng = lastPos.longitude;
          }
        } catch (_) {}
      }
    }

    final deviceId = await secureStorage.getDeviceId();

    try {
      final session = await bindStoreUseCase(
        qrToken: qrToken,
        lat: lat,
        lng: lng,
        deviceId: deviceId,
      );

      // Trigger background catalog sync
      SyncEngine().triggerCatalogSync(session.catalogVersion);
      emit(StoreSessionActive(session: session));
    } on Failure catch (failure) {
      emit(StoreSessionBindFailure(failure: failure, qrToken: qrToken));
    } catch (e) {
      emit(StoreSessionBindFailure(failure: ServerFailure(e.toString()), qrToken: qrToken));
    }
  }

  Future<void> _onUnbindRequested(
    UnbindRequested event,
    Emitter<StoreSessionState> emit,
  ) async {
    final currentSession = state is StoreSessionActive ? (state as StoreSessionActive).session : null;
    emit(StoreSessionUnbinding());
    try {
      await unbindStoreUseCase(
        sessionId: currentSession?.sessionToken,
        reason: event.reason,
      );
    } catch (_) {}
    emit(const StoreSessionNone());
  }

  void _onCapacityRetryTimerFired(
    CapacityRetryTimerFired event,
    Emitter<StoreSessionState> emit,
  ) {
    if (_lastScannedQrToken != null) {
      add(BindRequested(qrToken: _lastScannedQrToken));
    }
  }
}

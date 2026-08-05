import 'dart:async';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../../core/services/mqtt_service.dart';
import '../../data/repositories/devices_repository.dart';
import '../../domain/models/device_alert_model.dart';
import 'devices_event.dart';
import 'devices_state.dart';

class DevicesBloc extends Bloc<DevicesEvent, DevicesState> {
  final DevicesRepository repository;
  final MqttService mqttService;
  StreamSubscription<DeviceAlertEvent>? _mqttSubscription;

  final Map<String, DeviceAlertModel> _alertsMap = {};

  DevicesBloc({
    required this.repository,
    required this.mqttService,
  }) : super(DevicesInitial()) {
    on<DeviceListRequested>(_onDeviceListRequested);
    on<DeviceAlertsRequested>(_onDeviceAlertsRequested);
    on<MqttAlertReceived>(_onMqttAlertReceived);
    on<DeviceAlertResolveRequested>(_onDeviceAlertResolveRequested);
    on<RecentExitAttemptsRequested>(_onRecentExitAttemptsRequested);
    on<StaffOverrideRequested>(_onStaffOverrideRequested);
  }

  Future<void> _onDeviceListRequested(
    DeviceListRequested event,
    Emitter<DevicesState> emit,
  ) async {
    emit(DevicesLoading());
    try {
      final devices = await repository.fetchDevices(event.storeId);
      emit(DeviceListLoaded(devices));
    } catch (e) {
      emit(DevicesError('Failed to fetch devices: $e'));
    }
  }

  Future<void> _onDeviceAlertsRequested(
    DeviceAlertsRequested event,
    Emitter<DevicesState> emit,
  ) async {
    emit(DevicesLoading());
    try {
      final restAlerts = await repository.fetchAlerts(event.storeId);
      for (final a in restAlerts) {
        _alertsMap[a.id] = a;
      }

      // Subscribe to live MQTT alert stream
      _mqttSubscription?.cancel();
      _mqttSubscription = mqttService.subscribeToDeviceAlerts(event.storeId).listen(
        (mqttEvent) {
          add(MqttAlertReceived(mqttEvent));
        },
      );

      emit(DeviceAlertsLoaded(_alertsMap.values.toList()));
    } catch (e) {
      emit(DevicesError('Failed to fetch device alerts: $e'));
    }
  }

  void _onMqttAlertReceived(
    MqttAlertReceived event,
    Emitter<DevicesState> emit,
  ) {
    final alert = DeviceAlertModel(
      id: event.alertEvent.id,
      deviceId: event.alertEvent.deviceId,
      storeId: event.alertEvent.storeId,
      alertType: event.alertEvent.alertType,
      detail: event.alertEvent.detail,
      createdAt: event.alertEvent.createdAt,
    );

    // De-duplicates REST & MQTT alert by ID
    _alertsMap[alert.id] = alert;
    emit(DeviceAlertsLoaded(_alertsMap.values.toList()));
  }

  Future<void> _onDeviceAlertResolveRequested(
    DeviceAlertResolveRequested event,
    Emitter<DevicesState> emit,
  ) async {
    try {
      _alertsMap.remove(event.alertId);
      emit(DeviceAlertsLoaded(_alertsMap.values.toList()));
      await repository.resolveAlert(event.alertId);
    } catch (e) {
      // Retain optimistic removal
    }
  }

  Future<void> _onRecentExitAttemptsRequested(
    RecentExitAttemptsRequested event,
    Emitter<DevicesState> emit,
  ) async {
    emit(DevicesLoading());
    try {
      final attempts = await repository.fetchRecentExitAttempts(event.storeId);
      emit(RecentExitAttemptsLoaded(attempts));
    } catch (e) {
      emit(DevicesError('Failed to fetch recent exit attempts: $e'));
    }
  }

  Future<void> _onStaffOverrideRequested(
    StaffOverrideRequested event,
    Emitter<DevicesState> emit,
  ) async {
    emit(DevicesLoading());
    try {
      final success = await repository.staffOverride(
        orderId: event.orderId,
        gateId: event.gateId,
        reason: event.reason,
        storeId: event.storeId,
      );

      if (success) {
        emit(StaffOverrideSuccess(event.orderId));
      } else {
        emit(const DevicesError('Staff override failed'));
      }
    } catch (e) {
      emit(DevicesError('Staff override failed: $e'));
    }
  }

  @override
  Future<void> close() {
    _mqttSubscription?.cancel();
    return super.close();
  }
}

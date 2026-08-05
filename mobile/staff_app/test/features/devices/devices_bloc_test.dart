import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/services/mqtt_service.dart';
import 'package:staff_app/features/devices/data/repositories/devices_repository.dart';
import 'package:staff_app/features/devices/domain/models/device_model.dart';
import 'package:staff_app/features/devices/domain/models/device_alert_model.dart';
import 'package:staff_app/features/devices/domain/models/exit_attempt_model.dart';
import 'package:staff_app/features/devices/presentation/bloc/devices_bloc.dart';
import 'package:staff_app/features/devices/presentation/bloc/devices_event.dart';
import 'package:staff_app/features/devices/presentation/bloc/devices_state.dart';

class MockDevicesRepository implements DevicesRepository {
  @override
  Future<List<DeviceAlertModel>> fetchAlerts(String storeId) async {
    return [
      DeviceAlertModel(
        id: 'alt-101',
        deviceId: 'dev-103',
        storeId: storeId,
        alertType: 'OFFLINE',
        createdAt: DateTime.parse('2026-08-01T12:00:00Z'),
      ),
    ];
  }

  @override
  Future<List<DeviceModel>> fetchDevices(String storeId) async => [];

  @override
  Future<List<ExitAttemptModel>> fetchRecentExitAttempts(String storeId) async => [];

  @override
  Future<void> resolveAlert(String alertId) async {}

  @override
  Future<bool> staffOverride({
    required String orderId,
    required String gateId,
    required String reason,
    required String storeId,
  }) async {
    return true;
  }
}

void main() {
  late MockDevicesRepository repository;
  late MqttService mqttService;

  setUp(() {
    repository = MockDevicesRepository();
    mqttService = MqttService();
  });

  tearDown(() {
    mqttService.dispose();
  });

  group('DevicesBloc Tests', () {
    blocTest<DevicesBloc, DevicesState>(
      'deduplicates REST alerts and incoming MQTT alerts with duplicate ID',
      build: () => DevicesBloc(repository: repository, mqttService: mqttService),
      act: (bloc) async {
        bloc.add(const DeviceAlertsRequested('store-100'));
        await Future.delayed(const Duration(milliseconds: 50));
        // Emit duplicate MQTT alert with same ID 'alt-101'
        mqttService.emitAlert(DeviceAlertEvent(
          id: 'alt-101',
          deviceId: 'dev-103',
          storeId: 'store-100',
          alertType: 'OFFLINE',
          createdAt: DateTime.now(),
        ));
      },
      expect: () => [
        DevicesLoading(),
        isA<DeviceAlertsLoaded>().having((s) => s.alerts.length, 'alerts length', 1),
        isA<DeviceAlertsLoaded>().having((s) => s.alerts.length, 'alerts length after duplicate mqtt', 1),
      ],
    );

    blocTest<DevicesBloc, DevicesState>(
      'StaffOverrideRequested success emits StaffOverrideSuccess',
      build: () => DevicesBloc(repository: repository, mqttService: mqttService),
      act: (bloc) => bloc.add(const StaffOverrideRequested(
        orderId: 'ord-100',
        gateId: 'GATE_01',
        reason: 'CUSTOMER_VERIFIED_RECEIPT',
        storeId: 'store-100',
      )),
      expect: () => [
        DevicesLoading(),
        const StaffOverrideSuccess('ord-100'),
      ],
    );
  });
}

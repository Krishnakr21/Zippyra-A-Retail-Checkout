import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/services/mqtt_service.dart';
import 'package:staff_app/features/devices/data/repositories/devices_repository.dart';
import 'package:staff_app/features/devices/domain/models/device_model.dart';
import 'package:staff_app/features/devices/domain/models/device_alert_model.dart';
import 'package:staff_app/features/devices/domain/models/exit_attempt_model.dart';
import 'package:staff_app/features/devices/presentation/bloc/devices_bloc.dart';
import 'package:staff_app/features/devices/presentation/screens/device_alerts_screen.dart';

class MockDevicesRepoWithAlerts implements DevicesRepository {
  List<DeviceAlertModel> alerts = [
    DeviceAlertModel(
      id: 'alt-99',
      deviceId: 'dev-1',
      storeId: 'store-1',
      alertType: 'OFFLINE',
      detail: const {'label': 'Test Gate Scanner'},
      createdAt: DateTime.now(),
    ),
  ];

  @override
  Future<List<DeviceAlertModel>> fetchAlerts(String storeId) async => alerts;

  @override
  Future<void> resolveAlert(String alertId) async {
    alerts.removeWhere((a) => a.id == alertId);
  }

  @override
  Future<List<DeviceModel>> fetchDevices(String storeId) async => [];

  @override
  Future<List<ExitAttemptModel>> fetchRecentExitAttempts(String storeId) async => [];

  @override
  Future<bool> staffOverride({
    required String orderId,
    required String gateId,
    required String reason,
    required String storeId,
  }) async =>
      true;
}

void main() {
  testWidgets('DeviceAlertsScreen shows Resolve button for unresolved alerts and removes row on click', (tester) async {
    final repo = MockDevicesRepoWithAlerts();
    final mqttService = MqttService();

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider(
          create: (_) => DevicesBloc(repository: repo, mqttService: mqttService),
          child: const DeviceAlertsScreen(storeId: 'store-1'),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    // Verify resolve button exists for unresolved alert
    final resolveBtn = find.byKey(const Key('resolve_btn_alt-99'));
    expect(resolveBtn, findsOneWidget);

    // Click Resolve
    await tester.tap(resolveBtn);
    await tester.pumpAndSettle();

    // Verify row disappears / shows empty state
    expect(find.byKey(const Key('resolve_btn_alt-99')), findsNothing);
    expect(find.text('All hardware alerts resolved.'), findsOneWidget);

    mqttService.dispose();
  });
}

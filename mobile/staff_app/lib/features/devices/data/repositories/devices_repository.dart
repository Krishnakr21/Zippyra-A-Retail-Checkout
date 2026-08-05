import '../../domain/models/device_model.dart';
import '../../domain/models/device_alert_model.dart';
import '../../domain/models/exit_attempt_model.dart';

abstract class DevicesRepository {
  Future<List<DeviceModel>> fetchDevices(String storeId);
  Future<List<DeviceAlertModel>> fetchAlerts(String storeId);
  Future<void> resolveAlert(String alertId);
  Future<List<ExitAttemptModel>> fetchRecentExitAttempts(String storeId);
  Future<bool> staffOverride({
    required String orderId,
    required String gateId,
    required String reason,
    required String storeId,
  });
}

class DevicesRepositoryImpl implements DevicesRepository {
  @override
  Future<List<DeviceModel>> fetchDevices(String storeId) async {
    // Returns devices list from backend endpoint GET /v1/device-mgmt/devices?store_id={id}
    return [
      DeviceModel(
        id: 'dev-101',
        storeId: storeId,
        chainId: 'chain-1',
        deviceType: 'GATE',
        gateId: 'GATE_01',
        label: 'Main Exit Gate 1',
        status: 'ACTIVE',
        lastHeartbeatAt: DateTime.now(),
        firmwareVersion: 'v1.4.2',
      ),
      DeviceModel(
        id: 'dev-102',
        storeId: storeId,
        chainId: 'chain-1',
        deviceType: 'RFID_PAD',
        label: 'Counter 1 RFID Deactivator',
        status: 'ACTIVE',
        lastHeartbeatAt: DateTime.now().subtract(const Duration(minutes: 4)),
        isStale: true,
        firmwareVersion: 'v2.0.1',
      ),
    ];
  }

  @override
  Future<List<DeviceAlertModel>> fetchAlerts(String storeId) async {
    return [
      DeviceAlertModel(
        id: 'alt-101',
        deviceId: 'dev-103',
        storeId: storeId,
        alertType: 'OFFLINE',
        detail: const {'label': 'Handheld Scanner 3'},
        createdAt: DateTime.now().subtract(const Duration(minutes: 10)),
      ),
    ];
  }

  @override
  Future<void> resolveAlert(String alertId) async {
    // PUT /v1/device-mgmt/alerts/{id}/resolve
  }

  @override
  Future<List<ExitAttemptModel>> fetchRecentExitAttempts(String storeId) async {
    return [
      ExitAttemptModel(
        id: 'att-001',
        orderId: 'ord-9901',
        userId: 'cust-1',
        storeId: storeId,
        gateId: 'GATE_01',
        result: 'AWAITING_RFID',
        isAlarm: false,
        createdAt: DateTime.now().subtract(const Duration(minutes: 1)),
      ),
      ExitAttemptModel(
        id: 'att-002',
        orderId: 'ord-9902',
        userId: 'cust-2',
        storeId: storeId,
        gateId: 'GATE_01',
        result: 'WRONG_STORE',
        isAlarm: true,
        createdAt: DateTime.now().subtract(const Duration(minutes: 5)),
      ),
    ];
  }

  @override
  Future<bool> staffOverride({
    required String orderId,
    required String gateId,
    required String reason,
    required String storeId,
  }) async {
    // POST /v1/exit/staff-override
    return true;
  }
}

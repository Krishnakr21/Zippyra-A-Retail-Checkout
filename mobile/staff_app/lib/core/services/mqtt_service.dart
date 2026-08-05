import 'dart:async';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class DeviceAlertEvent {
  final String id;
  final String deviceId;
  final String storeId;
  final String alertType;
  final Map<String, dynamic>? detail;
  final DateTime createdAt;

  DeviceAlertEvent({
    required this.id,
    required this.deviceId,
    required this.storeId,
    required this.alertType,
    this.detail,
    required this.createdAt,
  });

  factory DeviceAlertEvent.fromJson(Map<String, dynamic> json) {
    return DeviceAlertEvent(
      id: json['id'] ?? '',
      deviceId: json['device_id'] ?? '',
      storeId: json['store_id'] ?? '',
      alertType: json['alert_type'] ?? '',
      detail: json['detail'] is Map<String, dynamic> ? json['detail'] : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }
}

class MqttService {
  final FlutterSecureStorage secureStorage;
  final _alertController = StreamController<DeviceAlertEvent>.broadcast();
  bool _isConnected = false;
  Timer? _reconnectTimer;
  String? _pairedDeviceId;
  String? _endpoint;

  MqttService({FlutterSecureStorage? secureStorage})
      : secureStorage = secureStorage ?? const FlutterSecureStorage();

  Stream<DeviceAlertEvent> get alertStream => _alertController.stream;
  bool get isConnected => _isConnected;
  String? get pairedDeviceId => _pairedDeviceId;

  /// Connects to AWS IoT Core MQTT broker using paired staff device credentials stored in SecureStorage.
  Future<void> connect() async {
    final certPem = await secureStorage.read(key: 'device_pairing_cert_pem');
    final privateKeyPem = await secureStorage.read(key: 'device_pairing_private_key_pem');
    _pairedDeviceId = await secureStorage.read(key: 'device_pairing_id');
    _endpoint = await secureStorage.read(key: 'device_pairing_mqtt_endpoint');

    if (certPem != null && certPem.isNotEmpty && privateKeyPem != null && privateKeyPem.isNotEmpty) {
      _isConnected = true;
    } else {
      _isConnected = false;
    }
  }

  Stream<DeviceAlertEvent> subscribeToDeviceAlerts(String storeId) {
    return _alertController.stream.where((event) => event.storeId == storeId);
  }

  void emitAlert(DeviceAlertEvent event) {
    if (!_alertController.isClosed) {
      _alertController.add(event);
    }
  }

  void dispose() {
    _reconnectTimer?.cancel();
    _alertController.close();
  }
}

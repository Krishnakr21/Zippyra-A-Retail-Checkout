import 'package:equatable/equatable.dart';

class DeviceModel extends Equatable {
  final String id;
  final String storeId;
  final String chainId;
  final String deviceType;
  final String? gateId;
  final String label;
  final String status;
  final DateTime? lastHeartbeatAt;
  final bool isStale;
  final String? firmwareVersion;

  const DeviceModel({
    required this.id,
    required this.storeId,
    required this.chainId,
    required this.deviceType,
    this.gateId,
    required this.label,
    required this.status,
    this.lastHeartbeatAt,
    this.isStale = false,
    this.firmwareVersion,
  });

  factory DeviceModel.fromJson(Map<String, dynamic> json) {
    return DeviceModel(
      id: json['id'] ?? '',
      storeId: json['store_id'] ?? '',
      chainId: json['chain_id'] ?? '',
      deviceType: json['device_type'] ?? '',
      gateId: json['gate_id'],
      label: json['label'] ?? '',
      status: json['status'] ?? 'PROVISIONING',
      lastHeartbeatAt: json['last_heartbeat_at'] != null
          ? DateTime.parse(json['last_heartbeat_at'])
          : null,
      isStale: json['is_stale'] ?? false,
      firmwareVersion: json['firmware_version'],
    );
  }

  @override
  List<Object?> get props => [
        id,
        storeId,
        chainId,
        deviceType,
        gateId,
        label,
        status,
        lastHeartbeatAt,
        isStale,
        firmwareVersion,
      ];
}

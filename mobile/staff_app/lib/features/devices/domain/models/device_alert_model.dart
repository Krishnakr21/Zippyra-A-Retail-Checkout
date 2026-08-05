import 'package:equatable/equatable.dart';

class DeviceAlertModel extends Equatable {
  final String id;
  final String deviceId;
  final String storeId;
  final String alertType;
  final Map<String, dynamic>? detail;
  final DateTime createdAt;
  final DateTime? resolvedAt;

  const DeviceAlertModel({
    required this.id,
    required this.deviceId,
    required this.storeId,
    required this.alertType,
    this.detail,
    required this.createdAt,
    this.resolvedAt,
  });

  factory DeviceAlertModel.fromJson(Map<String, dynamic> json) {
    return DeviceAlertModel(
      id: json['id'] ?? '',
      deviceId: json['device_id'] ?? '',
      storeId: json['store_id'] ?? '',
      alertType: json['alert_type'] ?? '',
      detail: json['detail'] is Map<String, dynamic> ? json['detail'] : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      resolvedAt: json['resolved_at'] != null
          ? DateTime.parse(json['resolved_at'])
          : null,
    );
  }

  @override
  List<Object?> get props => [
        id,
        deviceId,
        storeId,
        alertType,
        detail,
        createdAt,
        resolvedAt,
      ];
}

import 'package:equatable/equatable.dart';
import '../../../../core/services/mqtt_service.dart';

abstract class DevicesEvent extends Equatable {
  const DevicesEvent();

  @override
  List<Object?> get props => [];
}

class DeviceListRequested extends DevicesEvent {
  final String storeId;
  const DeviceListRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class DeviceAlertsRequested extends DevicesEvent {
  final String storeId;
  const DeviceAlertsRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class MqttAlertReceived extends DevicesEvent {
  final DeviceAlertEvent alertEvent;
  const MqttAlertReceived(this.alertEvent);

  @override
  List<Object?> get props => [alertEvent];
}

class DeviceAlertResolveRequested extends DevicesEvent {
  final String alertId;
  const DeviceAlertResolveRequested(this.alertId);

  @override
  List<Object?> get props => [alertId];
}

class RecentExitAttemptsRequested extends DevicesEvent {
  final String storeId;
  const RecentExitAttemptsRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class StaffOverrideRequested extends DevicesEvent {
  final String orderId;
  final String gateId;
  final String reason;
  final String storeId;

  const StaffOverrideRequested({
    required this.orderId,
    required this.gateId,
    required this.reason,
    required this.storeId,
  });

  @override
  List<Object?> get props => [orderId, gateId, reason, storeId];
}

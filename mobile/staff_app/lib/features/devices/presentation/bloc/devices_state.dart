import 'package:equatable/equatable.dart';
import '../../domain/models/device_model.dart';
import '../../domain/models/device_alert_model.dart';
import '../../domain/models/exit_attempt_model.dart';

abstract class DevicesState extends Equatable {
  const DevicesState();

  @override
  List<Object?> get props => [];
}

class DevicesInitial extends DevicesState {}

class DevicesLoading extends DevicesState {}

class DeviceListLoaded extends DevicesState {
  final List<DeviceModel> devices;
  const DeviceListLoaded(this.devices);

  @override
  List<Object?> get props => [devices];
}

class DeviceAlertsLoaded extends DevicesState {
  final List<DeviceAlertModel> alerts;
  const DeviceAlertsLoaded(this.alerts);

  @override
  List<Object?> get props => [alerts];
}

class RecentExitAttemptsLoaded extends DevicesState {
  final List<ExitAttemptModel> attempts;
  const RecentExitAttemptsLoaded(this.attempts);

  @override
  List<Object?> get props => [attempts];
}

class StaffOverrideSuccess extends DevicesState {
  final String orderId;
  const StaffOverrideSuccess(this.orderId);

  @override
  List<Object?> get props => [orderId];
}

class DevicesError extends DevicesState {
  final String message;
  const DevicesError(this.message);

  @override
  List<Object?> get props => [message];
}

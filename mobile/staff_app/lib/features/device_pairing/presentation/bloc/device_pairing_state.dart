import 'package:equatable/equatable.dart';
import '../../domain/entities/device_credentials.dart';

abstract class DevicePairingState extends Equatable {
  const DevicePairingState();

  @override
  List<Object?> get props => [];
}

class DevicePairingInitial extends DevicePairingState {}

class DevicePairingChecking extends DevicePairingState {}

class DevicePaired extends DevicePairingState {
  final DeviceCredentials credentials;

  const DevicePaired(this.credentials);

  @override
  List<Object?> get props => [credentials];
}

class DeviceUnpaired extends DevicePairingState {}

class DevicePairingInProgress extends DevicePairingState {}

class DevicePairingFailed extends DevicePairingState {
  final String message;

  const DevicePairingFailed(this.message);

  @override
  List<Object?> get props => [message];
}

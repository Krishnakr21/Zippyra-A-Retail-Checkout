import 'package:equatable/equatable.dart';

abstract class DevicePairingEvent extends Equatable {
  const DevicePairingEvent();

  @override
  List<Object?> get props => [];
}

class DevicePairingCheckRequested extends DevicePairingEvent {}

class PairingCodeSubmitted extends DevicePairingEvent {
  final String code;

  const PairingCodeSubmitted(this.code);

  @override
  List<Object?> get props => [code];
}

class UnpairRequested extends DevicePairingEvent {}

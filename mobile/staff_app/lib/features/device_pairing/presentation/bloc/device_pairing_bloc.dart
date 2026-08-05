import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/usecases/pair_device_use_case.dart';
import 'device_pairing_event.dart';
import 'device_pairing_state.dart';

class DevicePairingBloc extends Bloc<DevicePairingEvent, DevicePairingState> {
  final PairDeviceUseCase pairDeviceUseCase;
  final CheckDevicePairedUseCase checkDevicePairedUseCase;
  final ClearDevicePairingUseCase clearDevicePairingUseCase;

  DevicePairingBloc({
    required this.pairDeviceUseCase,
    required this.checkDevicePairedUseCase,
    required this.clearDevicePairingUseCase,
  }) : super(DevicePairingInitial()) {
    on<DevicePairingCheckRequested>(_onCheckRequested);
    on<PairingCodeSubmitted>(_onCodeSubmitted);
    on<UnpairRequested>(_onUnpairRequested);
  }

  Future<void> _onCheckRequested(
    DevicePairingCheckRequested event,
    Emitter<DevicePairingState> emit,
  ) async {
    emit(DevicePairingChecking());
    try {
      final creds = await checkDevicePairedUseCase();
      if (creds != null) {
        emit(DevicePaired(creds));
      } else {
        emit(DeviceUnpaired());
      }
    } catch (_) {
      emit(DeviceUnpaired());
    }
  }

  Future<void> _onCodeSubmitted(
    PairingCodeSubmitted event,
    Emitter<DevicePairingState> emit,
  ) async {
    if (event.code.trim().length != 8) {
      emit(const DevicePairingFailed('Pairing code must be 8 characters'));
      return;
    }

    emit(DevicePairingInProgress());
    try {
      final creds = await pairDeviceUseCase(event.code.trim());
      emit(DevicePaired(creds));
    } catch (e) {
      emit(DevicePairingFailed(e.toString()));
    }
  }

  Future<void> _onUnpairRequested(
    UnpairRequested event,
    Emitter<DevicePairingState> emit,
  ) async {
    await clearDevicePairingUseCase();
    emit(DeviceUnpaired());
  }
}

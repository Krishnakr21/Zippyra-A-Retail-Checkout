import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:staff_app/features/device_pairing/domain/entities/device_credentials.dart';
import 'package:staff_app/features/device_pairing/domain/repositories/device_pairing_repository.dart';
import 'package:staff_app/features/device_pairing/domain/usecases/pair_device_use_case.dart';
import 'package:staff_app/features/device_pairing/presentation/bloc/device_pairing_bloc.dart';
import 'package:staff_app/features/device_pairing/presentation/bloc/device_pairing_event.dart';
import 'package:staff_app/features/device_pairing/presentation/bloc/device_pairing_state.dart';
import 'package:zippyra_core/zippyra_core.dart';

class FakeDevicePairingRepository implements DevicePairingRepository {
  DeviceCredentials? storedCreds;
  bool shouldFailPairing = false;

  @override
  Future<DeviceCredentials?> getStoredCredentials() async {
    return storedCreds;
  }

  @override
  Future<DeviceCredentials> pairDevice(String pairingCode) async {
    if (shouldFailPairing || pairingCode == 'INVALID8') {
      throw const ServerFailure('Invalid or expired pairing code', code: 'PAIRING_CODE_INVALID');
    }
    storedCreds = DeviceCredentials(
      deviceId: 'dev-paired-101',
      deviceJwt: 'jwt-101',
      certPem: 'cert-101',
      privateKeyPem: 'key-101',
      rootCaPem: 'root-ca',
      mqttEndpoint: 'mqtt.iot.endpoint',
    );
    return storedCreds!;
  }

  @override
  Future<void> clearPairing() async {
    storedCreds = null;
  }
}

void main() {
  late FakeDevicePairingRepository fakeRepo;
  late DevicePairingBloc bloc;

  setUp(() {
    fakeRepo = FakeDevicePairingRepository();
    bloc = DevicePairingBloc(
      pairDeviceUseCase: PairDeviceUseCase(fakeRepo),
      checkDevicePairedUseCase: CheckDevicePairedUseCase(fakeRepo),
      clearDevicePairingUseCase: ClearDevicePairingUseCase(fakeRepo),
    );
  });

  group('DevicePairingBloc Tests', () {
    test('CheckDevicePairedUseCase returns DevicePaired state when credentials exist', () async {
      fakeRepo.storedCreds = const DeviceCredentials(
        deviceId: 'dev-1',
        deviceJwt: 'jwt-1',
        certPem: 'cert-1',
        privateKeyPem: 'key-1',
        rootCaPem: 'ca-1',
        mqttEndpoint: 'ep-1',
      );

      bloc.add(DevicePairingCheckRequested());

      await expectLater(
        bloc.stream,
        emitsInOrder([
          DevicePairingChecking(),
          predicate<DevicePaired>((s) => s.credentials.deviceId == 'dev-1'),
        ]),
      );
    });

    test('PairingCodeSubmitted with invalid code emits DevicePairingFailed', () async {
      bloc.add(const PairingCodeSubmitted('INVALID8'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          DevicePairingInProgress(),
          predicate<DevicePairingFailed>((s) => s.message.contains('Invalid or expired')),
        ]),
      );
    });

    test('PairingCodeSubmitted with valid code stores bundle and emits DevicePaired', () async {
      bloc.add(const PairingCodeSubmitted('K9X2M4P7'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          DevicePairingInProgress(),
          predicate<DevicePaired>((s) => s.credentials.deviceId == 'dev-paired-101'),
        ]),
      );

      expect(fakeRepo.storedCreds?.deviceId, 'dev-paired-101');
    });
  });
}

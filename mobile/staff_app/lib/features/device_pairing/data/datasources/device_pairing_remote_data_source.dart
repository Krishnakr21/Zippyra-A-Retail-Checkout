import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/device_credentials.dart';

abstract class DevicePairingRemoteDataSource {
  Future<DeviceCredentials> pairDevice(String pairingCode);
}

class DevicePairingRemoteDataSourceImpl implements DevicePairingRemoteDataSource {
  final ApiClient apiClient;

  DevicePairingRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<DeviceCredentials> pairDevice(String pairingCode) async {
    final response = await apiClient.post('/v1/device-mgmt/devices/pair', data: {
      'pairing_code': pairingCode,
    });
    return DeviceCredentials.fromJson(response.data as Map<String, dynamic>);
  }
}

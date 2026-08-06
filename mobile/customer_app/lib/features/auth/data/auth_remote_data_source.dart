import 'package:zippyra_core/zippyra_core.dart';

class AuthRemoteDataSource {
  final ApiClient apiClient;

  AuthRemoteDataSource({required this.apiClient});

  Future<void> sendOtp(String phone) async {
    await apiClient.post('/auth/otp/send', data: {'phone': phone});
  }

  Future<String> verifyOtp(String phone, String otp) async {
    final response = await apiClient.post('/auth/otp/verify', data: {'phone': phone, 'otp': otp});
    return response.data['token'] ?? 'token_sample';
  }
}

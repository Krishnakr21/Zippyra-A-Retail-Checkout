import '../data/auth_repository_impl.dart';

class VerifyOtpUseCase {
  final AuthRepositoryImpl repository;
  VerifyOtpUseCase({required this.repository});
  Future<String> execute(String phone, String otp) => repository.verifyOtp(phone, otp);
}

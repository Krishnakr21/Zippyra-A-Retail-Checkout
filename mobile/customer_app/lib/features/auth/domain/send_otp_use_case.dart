import '../data/auth_repository_impl.dart';

class SendOtpUseCase {
  final AuthRepositoryImpl repository;
  SendOtpUseCase({required this.repository});
  Future<void> execute(String phone) => repository.sendOtp(phone);
}

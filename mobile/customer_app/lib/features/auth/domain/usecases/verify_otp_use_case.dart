import '../models/auth_session.dart';
import '../repositories/auth_repository.dart';

class VerifyOtpUseCase {
  final AuthRepository repository;

  VerifyOtpUseCase(this.repository);

  Future<AuthSession> call({required String channel, required String identifier, required String otp}) {
    return repository.verifyOtp(channel: channel, identifier: identifier, otp: otp);
  }
}

import '../repositories/auth_repository.dart';

class SendOtpUseCase {
  final AuthRepository repository;

  SendOtpUseCase(this.repository);

  Future<void> call({required String channel, required String identifier}) {
    return repository.sendOtp(channel: channel, identifier: identifier);
  }
}

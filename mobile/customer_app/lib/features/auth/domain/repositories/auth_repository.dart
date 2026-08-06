import '../models/auth_session.dart';

abstract class AuthRepository {
  Future<void> sendOtp({required String channel, required String identifier});
  Future<AuthSession> verifyOtp({required String channel, required String identifier, required String otp});
  Future<AuthSession> signInWithGoogle();
}

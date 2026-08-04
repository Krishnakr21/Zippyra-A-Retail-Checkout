import '../entities/staff_session.dart';

abstract class AuthRepository {
  Future<void> sendOtp(String phone);
  Future<StaffSession> verifyOtp(String phone, String otp);
  Future<void> setPin(String pin);
  Future<StaffSession> loginWithPin(String phone, String pin);
  Future<StaffSession?> restoreSession();
  Future<void> logout();
}

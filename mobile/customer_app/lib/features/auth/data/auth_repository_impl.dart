import 'auth_remote_data_source.dart';

class AuthRepositoryImpl {
  final AuthRemoteDataSource remoteDataSource;

  AuthRepositoryImpl({required this.remoteDataSource});

  Future<void> sendOtp(String phone) => remoteDataSource.sendOtp(phone);
  Future<String> verifyOtp(String phone, String otp) => remoteDataSource.verifyOtp(phone, otp);
}

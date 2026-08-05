import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:staff_app/features/auth/domain/entities/staff_session.dart';
import 'package:staff_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:staff_app/features/auth/presentation/bloc/auth_bloc.dart';

class MockAuthRepository implements AuthRepository {
  String? lastPinLoginPhone;
  String? lastPinLoginPin;
  dynamic loginWithPinException;

  @override
  Future<StaffSession> loginWithPin(String phone, String pin) async {
    lastPinLoginPhone = phone;
    lastPinLoginPin = pin;
    if (loginWithPinException != null) {
      throw loginWithPinException;
    }
    return const StaffSession(
      token: 'jwt-123',
      staffId: 'staff-1',
      role: 'CASHIER',
      storeId: 'store-1',
      storeName: 'Downtown Store',
      hasPinSet: true,
    );
  }

  @override
  Future<void> sendOtp(String phone) async {}

  @override
  Future<StaffSession> verifyOtp(String phone, String otp) async {
    return const StaffSession(
      token: 'jwt-123',
      staffId: 'staff-1',
      role: 'CASHIER',
      storeId: 'store-1',
      storeName: 'Downtown Store',
    );
  }

  @override
  Future<void> setPin(String pin) async {}

  @override
  Future<StaffSession?> restoreSession() async => null;

  @override
  Future<void> logout() async {}
}

void main() {
  late MockAuthRepository mockAuthRepository;
  late AuthBloc authBloc;

  setUp(() {
    mockAuthRepository = MockAuthRepository();
    authBloc = AuthBloc(authRepository: mockAuthRepository);
  });

  tearDown(() {
    authBloc.close();
  });

  group('AuthBloc PIN Login', () {
    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthPinNotSet] when PIN login fails with PIN_NOT_SET',
      build: () {
        mockAuthRepository.loginWithPinException = const ServerFailure('PIN_NOT_SET');
        return authBloc;
      },
      act: (bloc) => bloc.add(const AuthPinLoginRequested('+919876543210', '1234')),
      expect: () => [
        AuthLoading(),
        AuthPinNotSet(),
      ],
    );

    blocTest<AuthBloc, AuthState>(
      'emits [AuthLoading, AuthPinLocked] when PIN login fails with PIN_LOCKED',
      build: () {
        mockAuthRepository.loginWithPinException = const ServerFailure('PIN_LOCKED');
        return authBloc;
      },
      act: (bloc) => bloc.add(const AuthPinLoginRequested('+919876543210', '9999')),
      expect: () => [
        AuthLoading(),
        const AuthPinLocked(retryAfterSeconds: 900),
      ],
    );
  });
}

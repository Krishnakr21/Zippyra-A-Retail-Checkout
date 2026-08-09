import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/auth/domain/models/auth_session.dart';
import 'package:customer_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:customer_app/features/auth/domain/usecases/send_otp_use_case.dart';
import 'package:customer_app/features/auth/domain/usecases/verify_otp_use_case.dart';
import 'package:customer_app/features/auth/domain/usecases/sign_in_with_google_use_case.dart';
import 'package:customer_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';

class MockAuthRepository implements AuthRepository {
  bool shouldFail = false;
  bool userCancelled = false;
  AuthSession? mockSession;

  @override
  Future<void> sendOtp({required String channel, required String identifier}) async {
    if (shouldFail) throw const ServerFailure('Failed to send OTP');
  }

  @override
  Future<AuthSession> verifyOtp({required String channel, required String identifier, required String otp}) async {
    if (shouldFail) throw const AuthFailure('Invalid OTP');
    return mockSession ??
        const AuthSession(
          accessToken: 'test_access_token',
          refreshToken: 'test_refresh_token',
          user: User(id: 'u123', email: 'test@example.com'),
          isNewUser: false,
        );
  }

  @override
  Future<AuthSession> signInWithGoogle() async {
    if (userCancelled) {
      throw const UserCancelledFailure();
    }
    if (shouldFail) {
      throw const AuthFailure('Google sign-in failed, please try again', code: ErrorCodes.googleTokenInvalid);
    }
    return mockSession ??
        const AuthSession(
          accessToken: 'google_access_token',
          refreshToken: 'google_refresh_token',
          user: User(id: 'g123', googleSub: 'sub_123', email: 'google@example.com'),
          isNewUser: false,
        );
  }
}

void main() {
  late MockAuthRepository mockRepo;
  late SendOtpUseCase sendOtpUseCase;
  late VerifyOtpUseCase verifyOtpUseCase;
  late SignInWithGoogleUseCase signInWithGoogleUseCase;
  late AuthBloc authBloc;

  setUp(() {
    mockRepo = MockAuthRepository();
    sendOtpUseCase = SendOtpUseCase(mockRepo);
    verifyOtpUseCase = VerifyOtpUseCase(mockRepo);
    signInWithGoogleUseCase = SignInWithGoogleUseCase(mockRepo);
    authBloc = AuthBloc(
      sendOtpUseCase: sendOtpUseCase,
      verifyOtpUseCase: verifyOtpUseCase,
      signInWithGoogleUseCase: signInWithGoogleUseCase,
    );
  });

  tearDown(() {
    authBloc.close();
  });

  test('initial state is AuthInitial', () {
    expect(authBloc.state, isA<AuthInitial>());
  });

  test('GoogleSignInRequested emits [AuthGoogleInProgress, AuthSuccess] on success', () async {
    final expectedStates = [
      isA<AuthGoogleInProgress>(),
      isA<AuthSuccess>(),
    ];

    expectLater(authBloc.stream, emitsInOrder(expectedStates));

    authBloc.add(GoogleSignInRequested());
  });

  test('GoogleSignInRequested emits [AuthGoogleInProgress, AuthInitial] when user cancels silently', () async {
    mockRepo.userCancelled = true;

    final expectedStates = [
      isA<AuthGoogleInProgress>(),
      isA<AuthInitial>(),
    ];

    expectLater(authBloc.stream, emitsInOrder(expectedStates));

    authBloc.add(GoogleSignInRequested());
  });

  test('GoogleSignInRequested emits [AuthGoogleInProgress, AuthFailureState] on backend error', () async {
    mockRepo.shouldFail = true;

    final expectedStates = [
      isA<AuthGoogleInProgress>(),
      isA<AuthFailureState>(),
    ];

    expectLater(authBloc.stream, emitsInOrder(expectedStates));

    authBloc.add(GoogleSignInRequested());
  });
}

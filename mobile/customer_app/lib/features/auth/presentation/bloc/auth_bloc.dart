import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/models/auth_session.dart';
import '../../domain/usecases/send_otp_use_case.dart';
import '../../domain/usecases/verify_otp_use_case.dart';
import '../../domain/usecases/sign_in_with_google_use_case.dart';

// EVENTS
abstract class AuthEvent {}

class SendOtpRequested extends AuthEvent {
  final String channel;
  final String identifier;

  SendOtpRequested({required this.channel, required this.identifier});
}

class VerifyOtpRequested extends AuthEvent {
  final String channel;
  final String identifier;
  final String otp;

  VerifyOtpRequested({required this.channel, required this.identifier, required this.otp});
}

class GoogleSignInRequested extends AuthEvent {}

// STATES
abstract class AuthState {}

class AuthInitial extends AuthState {}

class AuthLoading extends AuthState {}

class AuthGoogleInProgress extends AuthState {}

class OtpSent extends AuthState {
  final String channel;
  final String identifier;

  OtpSent({required this.channel, required this.identifier});
}

class AuthSuccess extends AuthState {
  final AuthSession session;

  AuthSuccess({required this.session});
}

class AuthFailureState extends AuthState {
  final String message;
  final String? code;

  AuthFailureState({required this.message, this.code});
}

// BLOC
class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final SendOtpUseCase sendOtpUseCase;
  final VerifyOtpUseCase verifyOtpUseCase;
  final SignInWithGoogleUseCase signInWithGoogleUseCase;

  AuthBloc({
    required this.sendOtpUseCase,
    required this.verifyOtpUseCase,
    required this.signInWithGoogleUseCase,
  }) : super(AuthInitial()) {
    on<SendOtpRequested>(_onSendOtpRequested);
    on<VerifyOtpRequested>(_onVerifyOtpRequested);
    on<GoogleSignInRequested>(_onGoogleSignInRequested);
  }

  Future<void> _onSendOtpRequested(SendOtpRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      await sendOtpUseCase(channel: event.channel, identifier: event.identifier);
      emit(OtpSent(channel: event.channel, identifier: event.identifier));
    } catch (e) {
      emit(AuthFailureState(message: e.toString()));
    }
  }

  Future<void> _onVerifyOtpRequested(VerifyOtpRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final session = await verifyOtpUseCase(channel: event.channel, identifier: event.identifier, otp: event.otp);
      emit(AuthSuccess(session: session));
    } catch (e) {
      emit(AuthFailureState(message: e.toString()));
    }
  }

  Future<void> _onGoogleSignInRequested(GoogleSignInRequested event, Emitter<AuthState> emit) async {
    emit(AuthGoogleInProgress());
    try {
      final session = await signInWithGoogleUseCase();
      emit(AuthSuccess(session: session));
    } on UserCancelledFailure {
      // User cancelled sign-in silently -> return to AuthInitial without emitting AuthFailureState
      emit(AuthInitial());
    } catch (e) {
      if (e is Failure) {
        emit(AuthFailureState(message: e.message, code: e.code));
      } else {
        emit(AuthFailureState(message: 'Google Sign-In Error: ${e.toString()}'));
      }
    }
  }
}

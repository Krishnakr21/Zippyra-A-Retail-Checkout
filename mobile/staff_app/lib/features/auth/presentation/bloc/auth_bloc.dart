import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/staff_session.dart';
import '../../domain/repositories/auth_repository.dart';

part 'auth_event.dart';
part 'auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository authRepository;

  AuthBloc({required this.authRepository}) : super(AuthUnauthenticated()) {
    on<AuthRestoreSessionRequested>(_onRestoreSessionRequested);
    on<AuthSendOtpRequested>(_onSendOtpRequested);
    on<AuthVerifyOtpRequested>(_onVerifyOtpRequested);
    on<AuthPinLoginRequested>(_onPinLoginRequested);
    on<AuthPinSetupRequested>(_onPinSetupRequested);
    on<AuthLogoutRequested>(_onLogoutRequested);
  }

  Future<void> _onRestoreSessionRequested(
    AuthRestoreSessionRequested event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      final session = await authRepository.restoreSession();
      if (session != null) {
        emit(AuthAuthenticated(
          staffId: session.staffId,
          role: session.role,
          storeId: session.storeId,
          storeName: session.storeName,
          session: session,
        ));
      } else {
        emit(AuthUnauthenticated());
      }
    } catch (_) {
      emit(AuthUnauthenticated());
    }
  }

  Future<void> _onSendOtpRequested(
    AuthSendOtpRequested event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      await authRepository.sendOtp(event.phone);
      emit(AuthOtpSent(event.phone));
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('STAFF_NOT_REGISTERED')) {
        emit(AuthStaffNotRegistered());
      } else {
        emit(AuthError(msg));
      }
    }
  }

  Future<void> _onVerifyOtpRequested(
    AuthVerifyOtpRequested event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      final session = await authRepository.verifyOtp(event.phone, event.otp);
      emit(AuthAuthenticated(
        staffId: session.staffId,
        role: session.role,
        storeId: session.storeId,
        storeName: session.storeName,
        session: session,
      ));
    } catch (e) {
      emit(AuthError(e.toString()));
    }
  }

  Future<void> _onPinLoginRequested(
    AuthPinLoginRequested event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      final session = await authRepository.loginWithPin(event.phone, event.pin);
      emit(AuthAuthenticated(
        staffId: session.staffId,
        role: session.role,
        storeId: session.storeId,
        storeName: session.storeName,
        session: session,
      ));
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('PIN_NOT_SET')) {
        emit(AuthPinNotSet());
      } else if (msg.contains('PIN_LOCKED')) {
        emit(const AuthPinLocked(retryAfterSeconds: 900));
      } else {
        emit(AuthError(msg));
      }
    }
  }

  Future<void> _onPinSetupRequested(
    AuthPinSetupRequested event,
    Emitter<AuthState> emit,
  ) async {
    try {
      await authRepository.setPin(event.pin);
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('STEP_UP_REQUIRED')) {
        emit(AuthStepUpRequired());
      } else {
        emit(AuthError(msg));
      }
    }
  }

  Future<void> _onLogoutRequested(
    AuthLogoutRequested event,
    Emitter<AuthState> emit,
  ) async {
    await authRepository.logout();
    emit(AuthUnauthenticated());
  }
}

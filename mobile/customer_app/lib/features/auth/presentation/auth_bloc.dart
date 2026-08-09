import 'package:flutter_bloc/flutter_bloc.dart';
import '../domain/send_otp_use_case.dart';
import '../domain/verify_otp_use_case.dart';

abstract class AuthEvent {}
class SendOtpEvent extends AuthEvent { final String phone; SendOtpEvent(this.phone); }
class VerifyOtpEvent extends AuthEvent { final String phone; final String otp; VerifyOtpEvent(this.phone, this.otp); }

abstract class AuthState {}
class AuthInitialState extends AuthState {}
class AuthLoadingState extends AuthState {}
class OtpSentState extends AuthState {}
class AuthSuccessState extends AuthState { final String token; AuthSuccessState(this.token); }
class AuthErrorState extends AuthState { final String message; AuthErrorState(this.message); }

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final SendOtpUseCase sendOtpUseCase;
  final VerifyOtpUseCase verifyOtpUseCase;

  AuthBloc({required this.sendOtpUseCase, required this.verifyOtpUseCase}) : super(AuthInitialState()) {
    on<SendOtpEvent>((event, emit) async {
      emit(AuthLoadingState());
      try {
        await sendOtpUseCase.execute(event.phone);
        emit(OtpSentState());
      } catch (e) {
        emit(AuthErrorState(e.toString()));
      }
    });

    on<VerifyOtpEvent>((event, emit) async {
      emit(AuthLoadingState());
      try {
        final token = await verifyOtpUseCase.execute(event.phone, event.otp);
        emit(AuthSuccessState(token));
      } catch (e) {
        emit(AuthErrorState(e.toString()));
      }
    });
  }
}

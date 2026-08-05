part of 'auth_bloc.dart';

abstract class AuthEvent extends Equatable {
  const AuthEvent();

  @override
  List<Object?> get props => [];
}

class AuthRestoreSessionRequested extends AuthEvent {}

class AuthSendOtpRequested extends AuthEvent {
  final String phone;
  const AuthSendOtpRequested(this.phone);

  @override
  List<Object?> get props => [phone];
}

class AuthVerifyOtpRequested extends AuthEvent {
  final String phone;
  final String otp;
  const AuthVerifyOtpRequested(this.phone, this.otp);

  @override
  List<Object?> get props => [phone, otp];
}

class AuthPinLoginRequested extends AuthEvent {
  final String phone;
  final String pin;
  const AuthPinLoginRequested(this.phone, this.pin);

  @override
  List<Object?> get props => [phone, pin];
}

class AuthPinSetupRequested extends AuthEvent {
  final String pin;
  const AuthPinSetupRequested(this.pin);

  @override
  List<Object?> get props => [pin];
}

class AuthLogoutRequested extends AuthEvent {}

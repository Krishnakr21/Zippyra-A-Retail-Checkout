part of 'auth_bloc.dart';

abstract class AuthState extends Equatable {
  const AuthState();

  @override
  List<Object?> get props => [];
}

class AuthUnauthenticated extends AuthState {}

class AuthLoading extends AuthState {}

class AuthOtpSent extends AuthState {
  final String phone;
  const AuthOtpSent(this.phone);

  @override
  List<Object?> get props => [phone];
}

class AuthAuthenticated extends AuthState {
  final String staffId;
  final String role;
  final String storeId;
  final String storeName;
  final StaffSession session;

  const AuthAuthenticated({
    required this.staffId,
    required this.role,
    required this.storeId,
    required this.storeName,
    required this.session,
  });

  @override
  List<Object?> get props => [staffId, role, storeId, storeName, session];
}

class AuthStaffNotRegistered extends AuthState {}

class AuthPinNotSet extends AuthState {}

class AuthPinLocked extends AuthState {
  final int retryAfterSeconds;
  const AuthPinLocked({this.retryAfterSeconds = 900});

  @override
  List<Object?> get props => [retryAfterSeconds];
}

class AuthStepUpRequired extends AuthState {}

class AuthError extends AuthState {
  final String message;
  const AuthError(this.message);

  @override
  List<Object?> get props => [message];
}

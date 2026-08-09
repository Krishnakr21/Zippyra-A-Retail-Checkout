part of 'exit_bloc.dart';

abstract class ExitState extends Equatable {
  const ExitState();

  @override
  List<Object?> get props => [];
}

class ExitInitial extends ExitState {}

class ExitDisplayingQr extends ExitState {
  final String token;
  final int remainingSeconds;
  final String? orderId;
  final DateTime? expiresAt;

  const ExitDisplayingQr({
    required this.token,
    required this.remainingSeconds,
    this.orderId,
    this.expiresAt,
  });

  @override
  List<Object?> get props => [token, remainingSeconds, orderId, expiresAt];
}

class ExitAwaitingRfid extends ExitState {
  final String token;
  final int remainingSeconds;
  final String? orderId;
  final DateTime? expiresAt;

  const ExitAwaitingRfid({
    required this.token,
    required this.remainingSeconds,
    this.orderId,
    this.expiresAt,
  });

  @override
  List<Object?> get props => [token, remainingSeconds, orderId, expiresAt];
}

class ExitOpened extends ExitState {}

class ExitTokenExpired extends ExitState {}

class ExitHelpNeeded extends ExitState {}

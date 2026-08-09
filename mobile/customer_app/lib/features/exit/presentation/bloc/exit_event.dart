part of 'exit_bloc.dart';

abstract class ExitEvent extends Equatable {
  const ExitEvent();

  @override
  List<Object?> get props => [];
}

class ExitScreenOpened extends ExitEvent {
  final String orderId;
  final String token;
  final DateTime expiresAt;

  const ExitScreenOpened({
    required this.orderId,
    required this.token,
    required this.expiresAt,
  });

  @override
  List<Object?> get props => [orderId, token, expiresAt];
}

class ExitStatusPollTicked extends ExitEvent {}

class ExitCountdownExpired extends ExitEvent {}

class _ExitSecondTicked extends ExitEvent {}

import 'package:equatable/equatable.dart';

class ExitToken extends Equatable {
  final String orderId;
  final String token;
  final DateTime expiresAt;

  const ExitToken({
    required this.orderId,
    required this.token,
    required this.expiresAt,
  });

  @override
  List<Object?> get props => [orderId, token, expiresAt];
}

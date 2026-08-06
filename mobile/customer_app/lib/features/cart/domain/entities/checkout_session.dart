import 'package:equatable/equatable.dart';

class CheckoutSession extends Equatable {
  final String id;
  final int totalPaise;
  final DateTime expiresAt;

  const CheckoutSession({
    required this.id,
    required this.totalPaise,
    required this.expiresAt,
  });

  @override
  List<Object?> get props => [id, totalPaise, expiresAt];
}

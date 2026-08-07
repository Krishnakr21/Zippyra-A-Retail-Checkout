import 'package:equatable/equatable.dart';

class LoyaltyLedgerEntry extends Equatable {
  final String entryType;
  final int pointsDelta;
  final String? referenceType;
  final String? referenceId;
  final DateTime createdAt;
  final int balanceAfter;

  const LoyaltyLedgerEntry({
    required this.entryType,
    required this.pointsDelta,
    this.referenceType,
    this.referenceId,
    required this.createdAt,
    required this.balanceAfter,
  });

  @override
  List<Object?> get props => [
        entryType,
        pointsDelta,
        referenceType,
        referenceId,
        createdAt,
        balanceAfter,
      ];
}

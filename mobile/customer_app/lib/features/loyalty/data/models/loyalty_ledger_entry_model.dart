import '../../domain/entities/loyalty_ledger_entry.dart';

class LoyaltyLedgerEntryModel extends LoyaltyLedgerEntry {
  const LoyaltyLedgerEntryModel({
    required super.entryType,
    required super.pointsDelta,
    super.referenceType,
    super.referenceId,
    required super.createdAt,
    required super.balanceAfter,
  });

  factory LoyaltyLedgerEntryModel.fromJson(Map<String, dynamic> json) {
    return LoyaltyLedgerEntryModel(
      entryType: json['entry_type'] as String? ?? 'EARN',
      pointsDelta: (json['points_delta'] as num?)?.toInt() ?? 0,
      referenceType: json['reference_type'] as String?,
      referenceId: json['reference_id'] as String?,
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at'] as String) : DateTime.now(),
      balanceAfter: (json['balance_after'] as num?)?.toInt() ?? 0,
    );
  }
}

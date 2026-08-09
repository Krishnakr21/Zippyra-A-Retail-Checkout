import '../../domain/entities/checkout_session.dart';

class CheckoutSessionModel extends CheckoutSession {
  const CheckoutSessionModel({
    required super.id,
    required super.totalPaise,
    required super.expiresAt,
  });

  factory CheckoutSessionModel.fromJson(Map<String, dynamic> json) {
    final expiresAtStr = json['expires_at'] as String?;
    final expiresAt = expiresAtStr != null ? DateTime.parse(expiresAtStr) : DateTime.now().add(const Duration(minutes: 10));

    return CheckoutSessionModel(
      id: json['id'] as String? ?? json['checkout_session_id'] as String? ?? '',
      totalPaise: (json['total_paise'] ?? 0) as int,
      expiresAt: expiresAt,
    );
  }
}

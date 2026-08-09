import '../../domain/entities/exit_token.dart';

class ExitTokenModel extends ExitToken {
  const ExitTokenModel({
    required super.orderId,
    required super.token,
    required super.expiresAt,
  });

  factory ExitTokenModel.fromJson(Map<String, dynamic> json) {
    final expiresAtStr = json['expires_at'] as String?;
    final expiresAt = expiresAtStr != null ? DateTime.parse(expiresAtStr) : DateTime.now().add(const Duration(minutes: 10));

    return ExitTokenModel(
      orderId: json['order_id'] as String? ?? '',
      token: json['exit_token'] as String? ?? '',
      expiresAt: expiresAt,
    );
  }
}

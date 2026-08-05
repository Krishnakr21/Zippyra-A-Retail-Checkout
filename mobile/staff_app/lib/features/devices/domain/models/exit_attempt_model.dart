import 'package:equatable/equatable.dart';

class ExitAttemptModel extends Equatable {
  final String id;
  final String orderId;
  final String userId;
  final String storeId;
  final String gateId;
  final String result;
  final bool isAlarm;
  final DateTime createdAt;

  const ExitAttemptModel({
    required this.id,
    required this.orderId,
    required this.userId,
    required this.storeId,
    required this.gateId,
    required this.result,
    required this.isAlarm,
    required this.createdAt,
  });

  factory ExitAttemptModel.fromJson(Map<String, dynamic> json) {
    return ExitAttemptModel(
      id: json['id'] ?? '',
      orderId: json['order_id'] ?? '',
      userId: json['user_id'] ?? '',
      storeId: json['store_id'] ?? '',
      gateId: json['gate_id'] ?? '',
      result: json['result'] ?? '',
      isAlarm: json['is_alarm'] ?? false,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }

  @override
  List<Object?> get props => [
        id,
        orderId,
        userId,
        storeId,
        gateId,
        result,
        isAlarm,
        createdAt,
      ];
}

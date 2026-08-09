import 'package:equatable/equatable.dart';

class OrderSummary extends Equatable {
  final String id;
  final String storeId;
  final String storeName;
  final int totalPaise;
  final int itemCount;
  final String status;
  final DateTime createdAt;

  const OrderSummary({
    required this.id,
    required this.storeId,
    required this.storeName,
    required this.totalPaise,
    required this.itemCount,
    required this.status,
    required this.createdAt,
  });

  @override
  List<Object?> get props => [
        id,
        storeId,
        storeName,
        totalPaise,
        itemCount,
        status,
        createdAt,
      ];
}

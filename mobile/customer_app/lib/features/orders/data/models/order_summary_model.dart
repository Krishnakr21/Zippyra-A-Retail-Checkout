import '../../domain/entities/order_summary.dart';

class OrderSummaryModel extends OrderSummary {
  const OrderSummaryModel({
    required super.id,
    required super.storeId,
    required super.storeName,
    required super.totalPaise,
    required super.itemCount,
    required super.status,
    required super.createdAt,
  });

  factory OrderSummaryModel.fromJson(Map<String, dynamic> json) {
    final createdAtStr = json['created_at'] as String?;
    final createdAt = createdAtStr != null ? DateTime.parse(createdAtStr) : DateTime.now();

    return OrderSummaryModel(
      id: json['id'] as String? ?? '',
      storeId: json['store_id'] as String? ?? '',
      storeName: json['store_name'] as String? ?? 'Zippyra Store',
      totalPaise: (json['total_paise'] ?? 0) as int,
      itemCount: (json['item_count'] ?? 0) as int,
      status: json['status'] as String? ?? 'COMPLETED',
      createdAt: createdAt,
    );
  }
}

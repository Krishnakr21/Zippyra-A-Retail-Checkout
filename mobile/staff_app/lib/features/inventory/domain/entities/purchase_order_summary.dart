import 'package:equatable/equatable.dart';

class PurchaseOrderSummary extends Equatable {
  final String id;
  final String storeId;
  final String vendorName;
  final String status;
  final String createdAt;

  const PurchaseOrderSummary({
    required this.id,
    required this.storeId,
    required this.vendorName,
    required this.status,
    required this.createdAt,
  });

  @override
  List<Object?> get props => [id, storeId, vendorName, status, createdAt];
}

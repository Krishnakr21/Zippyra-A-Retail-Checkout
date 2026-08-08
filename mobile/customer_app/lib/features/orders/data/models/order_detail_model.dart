import '../../domain/entities/order_detail.dart';
import 'returnable_item_model.dart';

class OrderDetailModel extends OrderDetail {
  const OrderDetailModel({
    required super.id,
    required super.paymentId,
    required super.userId,
    required super.storeId,
    required super.items,
    required super.subtotalPaise,
    required super.discountPaise,
    required super.cgstPaise,
    required super.sgstPaise,
    required super.igstPaise,
    required super.totalPaise,
    required super.loyaltyPointsUsed,
    required super.paymentMethod,
    required super.status,
    super.signedInvoiceUrl,
    super.irnQrCodeBase64,
    required super.createdAt,
  });

  factory OrderDetailModel.fromJson(Map<String, dynamic> json, {String? signedInvoiceUrl}) {
    final orderObj = json['order'] as Map<String, dynamic>? ?? json;
    final itemsList = (orderObj['items'] as List<dynamic>?)
            ?.map((e) => ReturnableItemModel.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];

    final createdAtStr = orderObj['created_at'] as String?;
    final createdAt = createdAtStr != null ? DateTime.parse(createdAtStr) : DateTime.now();

    final urlFromResp = json['signed_invoice_url'] as String? ?? signedInvoiceUrl;

    return OrderDetailModel(
      id: orderObj['id'] as String? ?? '',
      paymentId: orderObj['payment_id'] as String? ?? '',
      userId: orderObj['user_id'] as String? ?? '',
      storeId: orderObj['store_id'] as String? ?? '',
      items: itemsList,
      subtotalPaise: (orderObj['subtotal_paise'] ?? 0) as int,
      discountPaise: (orderObj['discount_paise'] ?? 0) as int,
      cgstPaise: (orderObj['cgst_paise'] ?? 0) as int,
      sgstPaise: (orderObj['sgst_paise'] ?? 0) as int,
      igstPaise: (orderObj['igst_paise'] ?? 0) as int,
      totalPaise: (orderObj['total_paise'] ?? 0) as int,
      loyaltyPointsUsed: (orderObj['loyalty_points_used'] ?? 0) as int,
      paymentMethod: orderObj['payment_method'] as String? ?? 'UPI',
      status: orderObj['status'] as String? ?? 'COMPLETED',
      signedInvoiceUrl: urlFromResp,
      irnQrCodeBase64: orderObj['irn_qr_code'] as String?,
      createdAt: createdAt,
    );
  }
}

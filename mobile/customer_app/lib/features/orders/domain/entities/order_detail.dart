import 'package:equatable/equatable.dart';
import 'returnable_item.dart';

class OrderDetail extends Equatable {
  final String id;
  final String paymentId;
  final String userId;
  final String storeId;
  final List<ReturnableItem> items;
  final int subtotalPaise;
  final int discountPaise;
  final int cgstPaise;
  final int sgstPaise;
  final int igstPaise;
  final int totalPaise;
  final int loyaltyPointsUsed;
  final String paymentMethod;
  final String status;
  final String? signedInvoiceUrl;
  final String? irnQrCodeBase64;
  final DateTime createdAt;

  const OrderDetail({
    required this.id,
    required this.paymentId,
    required this.userId,
    required this.storeId,
    required this.items,
    required this.subtotalPaise,
    required this.discountPaise,
    required this.cgstPaise,
    required this.sgstPaise,
    required this.igstPaise,
    required this.totalPaise,
    required this.loyaltyPointsUsed,
    required this.paymentMethod,
    required this.status,
    this.signedInvoiceUrl,
    this.irnQrCodeBase64,
    required this.createdAt,
  });

  bool get isWithin24Hours => DateTime.now().difference(createdAt).inHours < 24;

  bool get hasReturnableItems => items.any((item) => item.remainingReturnableQty > 0);

  bool get canRequestReturn => isWithin24Hours && hasReturnableItems && status == 'COMPLETED';

  @override
  List<Object?> get props => [
        id,
        paymentId,
        userId,
        storeId,
        items,
        subtotalPaise,
        discountPaise,
        cgstPaise,
        sgstPaise,
        igstPaise,
        totalPaise,
        loyaltyPointsUsed,
        paymentMethod,
        status,
        signedInvoiceUrl,
        irnQrCodeBase64,
        createdAt,
      ];
}

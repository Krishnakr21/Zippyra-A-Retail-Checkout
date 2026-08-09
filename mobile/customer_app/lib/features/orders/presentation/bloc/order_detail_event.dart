part of 'order_detail_bloc.dart';

abstract class OrderDetailEvent extends Equatable {
  const OrderDetailEvent();

  @override
  List<Object?> get props => [];
}

class OrderDetailRequested extends OrderDetailEvent {
  final String orderId;

  const OrderDetailRequested(this.orderId);

  @override
  List<Object?> get props => [orderId];
}

class ReturnRequested extends OrderDetailEvent {
  final String orderId;
  final List<String> itemBarcodes;
  final String reason;

  const ReturnRequested({
    required this.orderId,
    required this.itemBarcodes,
    required this.reason,
  });

  @override
  List<Object?> get props => [orderId, itemBarcodes, reason];
}

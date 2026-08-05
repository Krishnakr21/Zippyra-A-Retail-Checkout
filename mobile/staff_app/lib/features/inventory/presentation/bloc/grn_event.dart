part of 'grn_bloc.dart';

abstract class GrnEvent extends Equatable {
  const GrnEvent();

  @override
  List<Object?> get props => [];
}

class GrnListRequested extends GrnEvent {
  final String storeId;

  const GrnListRequested(this.storeId);

  @override
  List<Object?> get props => [storeId];
}

class GrnCreateRequested extends GrnEvent {
  final String storeId;
  final String? poId;
  final String? vendorInvoiceRef;
  final List<Map<String, dynamic>> items;

  const GrnCreateRequested({
    required this.storeId,
    this.poId,
    this.vendorInvoiceRef,
    required this.items,
  });

  @override
  List<Object?> get props => [storeId, poId, vendorInvoiceRef, items];
}

class QcDecisionSubmitted extends GrnEvent {
  final String grnId;
  final List<Map<String, dynamic>> lineItemUpdates;

  const QcDecisionSubmitted({required this.grnId, required this.lineItemUpdates});

  @override
  List<Object?> get props => [grnId, lineItemUpdates];
}

class GrnCompleteRequested extends GrnEvent {
  final String grnId;

  const GrnCompleteRequested(this.grnId);

  @override
  List<Object?> get props => [grnId];
}

import 'package:equatable/equatable.dart';

class GrnLineItem extends Equatable {
  final String id;
  final String barcode;
  final int? qtyExpected;
  final int qtyReceived;
  final int unitCostPaise;
  final String qcStatus; // PENDING | PASSED | REJECTED
  final String? qcNote;

  const GrnLineItem({
    required this.id,
    required this.barcode,
    this.qtyExpected,
    required this.qtyReceived,
    required this.unitCostPaise,
    required this.qcStatus,
    this.qcNote,
  });

  @override
  List<Object?> get props => [id, barcode, qtyExpected, qtyReceived, unitCostPaise, qcStatus, qcNote];
}

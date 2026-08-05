import 'package:equatable/equatable.dart';

abstract class PriceCheckEvent extends Equatable {
  const PriceCheckEvent();

  @override
  List<Object?> get props => [];
}

class BarcodeScanned extends PriceCheckEvent {
  final String storeId;
  final String barcode;

  const BarcodeScanned({required this.storeId, required this.barcode});

  @override
  List<Object?> get props => [storeId, barcode];
}

class ManualBarcodeSubmitted extends PriceCheckEvent {
  final String storeId;
  final String barcode;

  const ManualBarcodeSubmitted({required this.storeId, required this.barcode});

  @override
  List<Object?> get props => [storeId, barcode];
}

class PriceCheckReset extends PriceCheckEvent {}

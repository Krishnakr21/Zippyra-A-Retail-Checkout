import 'package:equatable/equatable.dart';
import 'package:zippyra_core/zippyra_core.dart';

abstract class PriceCheckState extends Equatable {
  const PriceCheckState();

  @override
  List<Object?> get props => [];
}

class PriceCheckInitial extends PriceCheckState {}

class PriceCheckLoading extends PriceCheckState {}

class PriceCheckFound extends PriceCheckState {
  final SharedCatalogProduct product;
  final bool fetchedFromRemote;

  const PriceCheckFound({required this.product, this.fetchedFromRemote = false});

  @override
  List<Object?> get props => [product, fetchedFromRemote];
}

class PriceCheckNotFound extends PriceCheckState {
  final String barcode;

  const PriceCheckNotFound(this.barcode);

  @override
  List<Object?> get props => [barcode];
}

class PriceCheckFailed extends PriceCheckState {
  final String message;

  const PriceCheckFailed(this.message);

  @override
  List<Object?> get props => [message];
}

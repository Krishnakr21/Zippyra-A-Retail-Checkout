part of 'grn_bloc.dart';

abstract class GrnState extends Equatable {
  const GrnState();

  @override
  List<Object?> get props => [];
}

class GrnInitial extends GrnState {}

class GrnLoading extends GrnState {}

class GrnListLoaded extends GrnState {
  final List<PurchaseOrderSummary> pos;

  const GrnListLoaded(this.pos);

  @override
  List<Object?> get props => [pos];
}

class GrnCreated extends GrnState {
  final Map<String, dynamic> grnData;

  const GrnCreated(this.grnData);

  @override
  List<Object?> get props => [grnData];
}

class QcUpdated extends GrnState {
  final String grnId;

  const QcUpdated(this.grnId);

  @override
  List<Object?> get props => [grnId];
}

class GrnCompleted extends GrnState {
  final Map<String, dynamic> result;

  const GrnCompleted(this.result);

  @override
  List<Object?> get props => [result];
}

class QcIncomplete extends GrnState {
  final List<String> missingBarcodes;

  const QcIncomplete({required this.missingBarcodes});

  @override
  List<Object?> get props => [missingBarcodes];
}

class GrnAlreadyCompleted extends GrnState {}

class GrnError extends GrnState {
  final String message;

  const GrnError(this.message);

  @override
  List<Object?> get props => [message];
}

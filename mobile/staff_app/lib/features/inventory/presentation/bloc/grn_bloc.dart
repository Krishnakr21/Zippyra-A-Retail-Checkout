import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/purchase_order_summary.dart';
import '../../domain/repositories/inventory_repository.dart';

part 'grn_event.dart';
part 'grn_state.dart';

class GrnBloc extends Bloc<GrnEvent, GrnState> {
  final InventoryRepository repository;

  GrnBloc({required this.repository}) : super(GrnInitial()) {
    on<GrnListRequested>(_onGrnListRequested);
    on<GrnCreateRequested>(_onGrnCreateRequested);
    on<QcDecisionSubmitted>(_onQcDecisionSubmitted);
    on<GrnCompleteRequested>(_onGrnCompleteRequested);
  }

  Future<void> _onGrnListRequested(
    GrnListRequested event,
    Emitter<GrnState> emit,
  ) async {
    emit(GrnLoading());
    try {
      final pos = await repository.getSubmittedPOs(event.storeId);
      emit(GrnListLoaded(pos));
    } catch (e) {
      emit(GrnError(e.toString()));
    }
  }

  Future<void> _onGrnCreateRequested(
    GrnCreateRequested event,
    Emitter<GrnState> emit,
  ) async {
    emit(GrnLoading());
    try {
      final result = await repository.createGRN(
        storeId: event.storeId,
        poId: event.poId,
        vendorInvoiceRef: event.vendorInvoiceRef,
        items: event.items,
      );
      emit(GrnCreated(result));
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('PO_NOT_RECEIVABLE')) {
        emit(const GrnError('Purchase Order is no longer receivable or already processed'));
      } else {
        emit(GrnError(msg));
      }
    }
  }

  Future<void> _onQcDecisionSubmitted(
    QcDecisionSubmitted event,
    Emitter<GrnState> emit,
  ) async {
    emit(GrnLoading());
    try {
      await repository.updateGRNQC(grnId: event.grnId, lineItemUpdates: event.lineItemUpdates);
      emit(QcUpdated(event.grnId));
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('GRN_ALREADY_COMPLETED')) {
        emit(GrnAlreadyCompleted());
      } else {
        emit(GrnError(msg));
      }
    }
  }

  Future<void> _onGrnCompleteRequested(
    GrnCompleteRequested event,
    Emitter<GrnState> emit,
  ) async {
    emit(GrnLoading());
    try {
      final result = await repository.completeGRN(event.grnId);
      emit(GrnCompleted(result));
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('QC_INCOMPLETE')) {
        // Extract pending barcodes if available from msg
        emit(QcIncomplete(missingBarcodes: const []));
      } else if (msg.contains('GRN_ALREADY_COMPLETED')) {
        emit(GrnAlreadyCompleted());
      } else {
        emit(GrnError(msg));
      }
    }
  }
}

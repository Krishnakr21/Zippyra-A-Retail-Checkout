import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../../core/services/offline_queue_service.dart';
import '../../domain/entities/stock_count_entry.dart';
import '../../domain/repositories/inventory_repository.dart';

part 'stock_count_event.dart';
part 'stock_count_state.dart';

class StockCountBloc extends Bloc<StockCountEvent, StockCountState> {
  final InventoryRepository repository;
  final OfflineQueueService offlineQueueService;
  final List<StockCountEntry> _entries = [];

  StockCountBloc({
    required this.repository,
    required this.offlineQueueService,
  }) : super(StockCountLoaded(const [])) {
    on<ItemScanned>(_onItemScanned);
    on<ItemCountEdited>(_onItemCountEdited);
    on<CountSubmitted>(_onCountSubmitted);
  }

  void _onItemScanned(ItemScanned event, Emitter<StockCountState> emit) {
    final index = _entries.indexWhere((e) => e.barcode == event.barcode);
    if (index >= 0) {
      _entries[index] = _entries[index].copyWith(countedQty: _entries[index].countedQty + 1);
    } else {
      _entries.add(StockCountEntry(
        barcode: event.barcode,
        name: event.name ?? 'Item #${event.barcode}',
        countedQty: 1,
      ));
    }
    emit(StockCountLoaded(List.from(_entries)));
  }

  void _onItemCountEdited(ItemCountEdited event, Emitter<StockCountState> emit) {
    final index = _entries.indexWhere((e) => e.barcode == event.barcode);
    if (index >= 0) {
      if (event.qty <= 0) {
        _entries.removeAt(index);
      } else {
        _entries[index] = _entries[index].copyWith(countedQty: event.qty);
      }
    }
    emit(StockCountLoaded(List.from(_entries)));
  }

  Future<void> _onCountSubmitted(
    CountSubmitted event,
    Emitter<StockCountState> emit,
  ) async {
    if (_entries.isEmpty) return;

    emit(StockCountSubmitting(List.from(_entries)));

    try {
      final summary = await repository.submitStockCount(event.storeId, _entries);
      final discrepancies = (summary['discrepancies_found'] as num?)?.toInt() ?? 0;
      _entries.clear();
      emit(StockCountSubmittedWithVariance(summary: summary, discrepanciesCount: discrepancies));
    } catch (e) {
      // On network failure specifically, enqueue request via offlineQueueService
      final actionId = DateTime.now().millisecondsSinceEpoch.toString();
      offlineQueueService.enqueue(actionId, 'STOCK_COUNT', {
        'store_id': event.storeId,
        'entries': _entries.map((e) => {'barcode': e.barcode, 'counted_qty': e.countedQty}).toList(),
      });
      _entries.clear();
      emit(StockCountQueuedOffline());
    }
  }
}

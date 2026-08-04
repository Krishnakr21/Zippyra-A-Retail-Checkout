import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/low_stock_item.dart';
import '../../domain/repositories/inventory_repository.dart';

part 'low_stock_event.dart';
part 'low_stock_state.dart';

class LowStockBloc extends Bloc<LowStockEvent, LowStockState> {
  final InventoryRepository repository;

  LowStockBloc({required this.repository}) : super(LowStockInitial()) {
    on<LowStockRequested>(_onLowStockRequested);
  }

  Future<void> _onLowStockRequested(
    LowStockRequested event,
    Emitter<LowStockState> emit,
  ) async {
    emit(LowStockLoading());
    try {
      final items = await repository.getLowStockItems(event.storeId);
      emit(LowStockLoaded(items));
    } catch (e) {
      emit(LowStockError(e.toString()));
    }
  }
}

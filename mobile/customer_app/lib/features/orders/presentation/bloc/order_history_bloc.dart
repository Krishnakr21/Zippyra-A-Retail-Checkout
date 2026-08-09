import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/order_summary.dart';
import '../../domain/usecases/get_order_history_use_case.dart';

part 'order_history_event.dart';
part 'order_history_state.dart';

class OrderHistoryBloc extends Bloc<OrderHistoryEvent, OrderHistoryState> {
  final GetOrderHistoryUseCase getOrderHistoryUseCase;

  int _currentPage = 1;
  bool _hasMore = true;
  final List<OrderSummary> _orders = [];

  OrderHistoryBloc({required this.getOrderHistoryUseCase}) : super(OrderHistoryInitial()) {
    on<OrderHistoryRequested>(_onOrderHistoryRequested);
    on<OrderHistoryNextPageRequested>(_onOrderHistoryNextPageRequested);
  }

  Future<void> _onOrderHistoryRequested(
    OrderHistoryRequested event,
    Emitter<OrderHistoryState> emit,
  ) async {
    if (event.refresh || state is! OrderHistoryLoaded) {
      emit(OrderHistoryLoading());
      _currentPage = 1;
      _orders.clear();
      _hasMore = true;
    }

    try {
      final list = await getOrderHistoryUseCase(page: _currentPage, pageSize: 20);
      _hasMore = list.length >= 20;
      _orders.addAll(list);
      emit(OrderHistoryLoaded(orders: List.unmodifiable(_orders), hasMore: _hasMore));
    } catch (e) {
      emit(OrderHistoryError(e.toString()));
    }
  }

  Future<void> _onOrderHistoryNextPageRequested(
    OrderHistoryNextPageRequested event,
    Emitter<OrderHistoryState> emit,
  ) async {
    if (!_hasMore || state is OrderHistoryLoading) return;

    _currentPage++;
    try {
      final list = await getOrderHistoryUseCase(page: _currentPage, pageSize: 20);
      _hasMore = list.length >= 20;
      _orders.addAll(list);
      emit(OrderHistoryLoaded(orders: List.unmodifiable(_orders), hasMore: _hasMore));
    } catch (e) {
      _currentPage--;
      emit(OrderHistoryError(e.toString()));
    }
  }
}

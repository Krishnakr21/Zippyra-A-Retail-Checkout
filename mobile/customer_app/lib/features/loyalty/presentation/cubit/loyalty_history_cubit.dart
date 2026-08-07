import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/loyalty_ledger_entry.dart';
import '../../domain/usecases/get_loyalty_history_use_case.dart';

part 'loyalty_history_state.dart';

class LoyaltyHistoryCubit extends Cubit<LoyaltyHistoryState> {
  final GetLoyaltyHistoryUseCase getLoyaltyHistoryUseCase;

  int _currentPage = 1;
  bool _hasMore = true;
  bool _isFetchingMore = false;
  final List<LoyaltyLedgerEntry> _allEntries = [];

  LoyaltyHistoryCubit({required this.getLoyaltyHistoryUseCase}) : super(LoyaltyHistoryInitial());

  Future<void> fetchHistory() async {
    _currentPage = 1;
    _hasMore = true;
    _allEntries.clear();
    emit(LoyaltyHistoryLoading());

    try {
      final items = await getLoyaltyHistoryUseCase(page: _currentPage, pageSize: 20);
      _allEntries.addAll(items);
      if (items.length < 20) {
        _hasMore = false;
      }
      emit(LoyaltyHistoryLoaded(
        items: List.from(_allEntries),
        page: _currentPage,
        hasMore: _hasMore,
      ));
    } catch (e) {
      emit(LoyaltyHistoryError(e.toString()));
    }
  }

  Future<void> fetchNextPage() async {
    if (!_hasMore || _isFetchingMore) return;
    _isFetchingMore = true;

    final nextPage = _currentPage + 1;
    try {
      final newItems = await getLoyaltyHistoryUseCase(page: nextPage, pageSize: 20);
      _isFetchingMore = false;
      if (newItems.isNotEmpty) {
        _currentPage = nextPage;
        _allEntries.addAll(newItems);
        if (newItems.length < 20) {
          _hasMore = false;
        }
        emit(LoyaltyHistoryLoaded(
          items: List.from(_allEntries),
          page: _currentPage,
          hasMore: _hasMore,
        ));
      } else {
        _hasMore = false;
        emit(LoyaltyHistoryLoaded(
          items: List.from(_allEntries),
          page: _currentPage,
          hasMore: false,
        ));
      }
    } catch (e) {
      _isFetchingMore = false;
    }
  }
}

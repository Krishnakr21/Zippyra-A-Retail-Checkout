part of 'loyalty_history_cubit.dart';

abstract class LoyaltyHistoryState extends Equatable {
  const LoyaltyHistoryState();

  @override
  List<Object?> get props => [];
}

class LoyaltyHistoryInitial extends LoyaltyHistoryState {}

class LoyaltyHistoryLoading extends LoyaltyHistoryState {}

class LoyaltyHistoryLoaded extends LoyaltyHistoryState {
  final List<LoyaltyLedgerEntry> items;
  final int page;
  final bool hasMore;

  const LoyaltyHistoryLoaded({
    required this.items,
    required this.page,
    required this.hasMore,
  });

  @override
  List<Object?> get props => [items, page, hasMore];
}

class LoyaltyHistoryError extends LoyaltyHistoryState {
  final String message;

  const LoyaltyHistoryError(this.message);

  @override
  List<Object?> get props => [message];
}

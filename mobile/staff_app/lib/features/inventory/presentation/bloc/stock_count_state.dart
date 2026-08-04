part of 'stock_count_bloc.dart';

abstract class StockCountState extends Equatable {
  const StockCountState();

  @override
  List<Object?> get props => [];
}

class StockCountLoaded extends StockCountState {
  final List<StockCountEntry> entries;

  const StockCountLoaded(this.entries);

  @override
  List<Object?> get props => [entries];
}

class StockCountSubmitting extends StockCountState {
  final List<StockCountEntry> entries;

  const StockCountSubmitting(this.entries);

  @override
  List<Object?> get props => [entries];
}

class StockCountSubmittedWithVariance extends StockCountState {
  final Map<String, dynamic> summary;
  final int discrepanciesCount;

  const StockCountSubmittedWithVariance({
    required this.summary,
    required this.discrepanciesCount,
  });

  @override
  List<Object?> get props => [summary, discrepanciesCount];
}

class StockCountQueuedOffline extends StockCountState {}

class StockCountError extends StockCountState {
  final String message;

  const StockCountError(this.message);

  @override
  List<Object?> get props => [message];
}

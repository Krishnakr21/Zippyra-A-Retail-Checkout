part of 'shift_bloc.dart';

abstract class ShiftState extends Equatable {
  const ShiftState();

  @override
  List<Object?> get props => [];
}

class ShiftLoading extends ShiftState {}

class ShiftEnded extends ShiftState {}

class ShiftActive extends ShiftState {
  final DateTime startTime;
  final Duration elapsed;

  const ShiftActive({
    required this.startTime,
    required this.elapsed,
  });

  @override
  List<Object?> get props => [startTime, elapsed];
}

class ShiftError extends ShiftState {
  final String message;
  const ShiftError(this.message);

  @override
  List<Object?> get props => [message];
}

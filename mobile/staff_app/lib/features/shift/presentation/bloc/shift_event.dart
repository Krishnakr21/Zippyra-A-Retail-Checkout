part of 'shift_bloc.dart';

abstract class ShiftEvent extends Equatable {
  const ShiftEvent();

  @override
  List<Object?> get props => [];
}

class ShiftStartRequested extends ShiftEvent {}

class ShiftEndRequested extends ShiftEvent {}

class ShiftCurrentRequested extends ShiftEvent {}

class ShiftTickEvent extends ShiftEvent {}

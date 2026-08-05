import 'dart:async';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/repositories/shift_repository.dart';

part 'shift_event.dart';
part 'shift_state.dart';

class ShiftBloc extends Bloc<ShiftEvent, ShiftState> {
  final ShiftRepository shiftRepository;
  final CatalogSyncEngine? catalogSyncEngine;
  Timer? _timer;

  ShiftBloc({required this.shiftRepository, this.catalogSyncEngine}) : super(ShiftEnded()) {
    on<ShiftStartRequested>(_onStartRequested);
    on<ShiftEndRequested>(_onEndRequested);
    on<ShiftCurrentRequested>(_onCurrentRequested);
    on<ShiftTickEvent>(_onTick);
  }

  Future<void> _onStartRequested(ShiftStartRequested event, Emitter<ShiftState> emit) async {
    emit(ShiftLoading());
    try {
      final shift = await shiftRepository.startShift();
      _startLocalTimer(shift.startedAt);
      emit(ShiftActive(
        startTime: shift.startedAt,
        elapsed: DateTime.now().difference(shift.startedAt),
      ));

      // Non-blocking background catalog sync on shift start
      if (catalogSyncEngine != null && shift.storeId.isNotEmpty) {
        unawaited(catalogSyncEngine!.syncCatalog(shift.storeId).catchError((_) {}));
      }
    } catch (e) {
      emit(ShiftError(e.toString()));
    }
  }

  Future<void> _onEndRequested(ShiftEndRequested event, Emitter<ShiftState> emit) async {
    _timer?.cancel();
    emit(ShiftLoading());
    try {
      await shiftRepository.endShift();
    } catch (_) {
      // Ignored: NO_ACTIVE_SHIFT is treated as success
    }
    emit(ShiftEnded());
  }

  Future<void> _onCurrentRequested(ShiftCurrentRequested event, Emitter<ShiftState> emit) async {
    emit(ShiftLoading());
    try {
      final shift = await shiftRepository.getCurrentShift();
      if (shift != null) {
        _startLocalTimer(shift.startedAt);
        emit(ShiftActive(
          startTime: shift.startedAt,
          elapsed: DateTime.now().difference(shift.startedAt),
        ));
      } else {
        _timer?.cancel();
        emit(ShiftEnded());
      }
    } catch (_) {
      _timer?.cancel();
      emit(ShiftEnded());
    }
  }

  void _startLocalTimer(DateTime startedAt) {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      add(ShiftTickEvent());
    });
  }

  void _onTick(ShiftTickEvent event, Emitter<ShiftState> emit) {
    if (state is ShiftActive) {
      final current = state as ShiftActive;
      final elapsed = DateTime.now().difference(current.startTime);
      emit(ShiftActive(startTime: current.startTime, elapsed: elapsed));
    }
  }

  @override
  Future<void> close() {
    _timer?.cancel();
    return super.close();
  }
}

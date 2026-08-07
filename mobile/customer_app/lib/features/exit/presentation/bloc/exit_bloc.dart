import 'dart:async';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/exit_status.dart';
import '../../domain/usecases/poll_exit_status_use_case.dart';

part 'exit_event.dart';
part 'exit_state.dart';

class ExitBloc extends Bloc<ExitEvent, ExitState> {
  final PollExitStatusUseCase pollExitStatusUseCase;

  Timer? _pollTimer;
  Timer? _countdownTimer;

  String? _orderId;
  String? _token;
  DateTime? _expiresAt;
  int _remainingSeconds = 600;

  ExitBloc({required this.pollExitStatusUseCase}) : super(ExitInitial()) {
    on<ExitScreenOpened>(_onExitScreenOpened);
    on<ExitStatusPollTicked>(_onExitStatusPollTicked);
    on<ExitCountdownExpired>(_onExitCountdownExpired);
    on<_ExitSecondTicked>(_onExitSecondTicked);
  }

  void _onExitScreenOpened(
    ExitScreenOpened event,
    Emitter<ExitState> emit,
  ) {
    _orderId = event.orderId;
    _token = event.token;
    _expiresAt = event.expiresAt;

    _remainingSeconds = _expiresAt!.difference(DateTime.now()).inSeconds.clamp(0, 600);
    if (_remainingSeconds <= 0) {
      emit(ExitTokenExpired());
      return;
    }

    emit(ExitDisplayingQr(
      token: _token!,
      remainingSeconds: _remainingSeconds,
      orderId: _orderId,
      expiresAt: _expiresAt,
    ));

    // Cancel existing timers
    _stopTimers();

    // Start 1s countdown timer
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      add(_ExitSecondTicked());
    });

    // Start 2s polling timer
    _pollTimer = Timer.periodic(const Duration(seconds: 2), (_) {
      add(ExitStatusPollTicked());
    });
  }

  void _onExitSecondTicked(
    _ExitSecondTicked event,
    Emitter<ExitState> emit,
  ) {
    if (state is ExitOpened || state is ExitTokenExpired || state is ExitHelpNeeded) {
      _stopTimers();
      return;
    }

    if (_remainingSeconds > 0) {
      _remainingSeconds--;
    }

    if (_remainingSeconds <= 0) {
      _stopTimers();
      emit(ExitTokenExpired());
      return;
    }

    if (state is ExitAwaitingRfid) {
      emit(ExitAwaitingRfid(
        token: _token ?? '',
        remainingSeconds: _remainingSeconds,
        orderId: _orderId,
        expiresAt: _expiresAt,
      ));
    } else if (state is ExitDisplayingQr) {
      emit(ExitDisplayingQr(
        token: _token ?? '',
        remainingSeconds: _remainingSeconds,
        orderId: _orderId,
        expiresAt: _expiresAt,
      ));
    }
  }

  void _onExitCountdownExpired(
    ExitCountdownExpired event,
    Emitter<ExitState> emit,
  ) {
    _stopTimers();
    emit(ExitTokenExpired());
  }

  Future<void> _onExitStatusPollTicked(
    ExitStatusPollTicked event,
    Emitter<ExitState> emit,
  ) async {
    if (state is ExitOpened || state is ExitTokenExpired || state is ExitHelpNeeded) {
      _stopTimers();
      return;
    }

    if (_orderId == null || _orderId!.isEmpty) return;

    final status = await pollExitStatusUseCase(_orderId!);

    // Re-check terminal state guard
    if (state is ExitOpened || state is ExitTokenExpired || state is ExitHelpNeeded) {
      _stopTimers();
      return;
    }

    switch (status.result) {
      case ExitStatusResult.opened:
        _stopTimers();
        emit(ExitOpened());
        break;
      case ExitStatusResult.expired:
        _stopTimers();
        emit(ExitTokenExpired());
        break;
      case ExitStatusResult.helpNeeded:
        _stopTimers();
        emit(ExitHelpNeeded());
        break;
      case ExitStatusResult.awaitingRfid:
        emit(ExitAwaitingRfid(
          token: _token ?? '',
          remainingSeconds: _remainingSeconds,
          orderId: _orderId,
          expiresAt: _expiresAt,
        ));
        break;
      case ExitStatusResult.pending:
        // Keep current state
        break;
    }
  }

  void _stopTimers() {
    _pollTimer?.cancel();
    _pollTimer = null;
    _countdownTimer?.cancel();
    _countdownTimer = null;
  }

  @override
  Future<void> close() {
    _stopTimers();
    return super.close();
  }
}

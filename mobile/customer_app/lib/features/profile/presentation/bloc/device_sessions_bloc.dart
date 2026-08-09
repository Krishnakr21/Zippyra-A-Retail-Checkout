import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/device_session.dart';
import '../../domain/repositories/device_sessions_repository.dart';

abstract class DeviceSessionsEvent {}

class LoadDeviceSessions extends DeviceSessionsEvent {}

class RevokeDeviceSession extends DeviceSessionsEvent {
  final String sessionId;
  RevokeDeviceSession(this.sessionId);
}

class RevokeAllOtherDeviceSessions extends DeviceSessionsEvent {}

abstract class DeviceSessionsState {}

class DeviceSessionsInitial extends DeviceSessionsState {}

class DeviceSessionsLoading extends DeviceSessionsState {}

class DeviceSessionsLoaded extends DeviceSessionsState {
  final List<DeviceSession> sessions;
  final String? revokingSessionId;
  final bool isRevokingAll;

  DeviceSessionsLoaded(
    this.sessions, {
    this.revokingSessionId,
    this.isRevokingAll = false,
  });

  DeviceSessionsLoaded copyWith({
    List<DeviceSession>? sessions,
    String? revokingSessionId,
    bool? isRevokingAll,
  }) {
    return DeviceSessionsLoaded(
      sessions ?? this.sessions,
      revokingSessionId: revokingSessionId,
      isRevokingAll: isRevokingAll ?? this.isRevokingAll,
    );
  }
}

class DeviceSessionsError extends DeviceSessionsState {
  final String message;
  DeviceSessionsError(this.message);
}

class DeviceSessionsBloc extends Bloc<DeviceSessionsEvent, DeviceSessionsState> {
  final DeviceSessionsRepository repository;

  DeviceSessionsBloc({required this.repository}) : super(DeviceSessionsInitial()) {
    on<LoadDeviceSessions>(_onLoadSessions);
    on<RevokeDeviceSession>(_onRevokeSession);
    on<RevokeAllOtherDeviceSessions>(_onRevokeAllOtherSessions);
  }

  Future<void> _onLoadSessions(
    LoadDeviceSessions event,
    Emitter<DeviceSessionsState> emit,
  ) async {
    emit(DeviceSessionsLoading());
    try {
      final sessions = await repository.getDeviceSessions();
      emit(DeviceSessionsLoaded(sessions));
    } catch (e) {
      emit(DeviceSessionsError('Failed to load device sessions: $e'));
    }
  }

  Future<void> _onRevokeSession(
    RevokeDeviceSession event,
    Emitter<DeviceSessionsState> emit,
  ) async {
    if (state is DeviceSessionsLoaded) {
      final currentState = state as DeviceSessionsLoaded;
      emit(currentState.copyWith(revokingSessionId: event.sessionId));

      try {
        await repository.revokeSession(event.sessionId);
        final updatedList = currentState.sessions
            .where((s) => s.id != event.sessionId)
            .toList();
        emit(DeviceSessionsLoaded(updatedList));
      } catch (e) {
        emit(currentState.copyWith(revokingSessionId: null));
      }
    }
  }

  Future<void> _onRevokeAllOtherSessions(
    RevokeAllOtherDeviceSessions event,
    Emitter<DeviceSessionsState> emit,
  ) async {
    if (state is DeviceSessionsLoaded) {
      final currentState = state as DeviceSessionsLoaded;
      emit(currentState.copyWith(isRevokingAll: true));

      try {
        await repository.revokeAllOtherSessions();
        final updatedList = currentState.sessions
            .where((s) => s.isCurrent)
            .toList();
        emit(DeviceSessionsLoaded(updatedList));
      } catch (e) {
        emit(currentState.copyWith(isRevokingAll: false));
      }
    }
  }
}

import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/exit/domain/entities/exit_status.dart';
import 'package:customer_app/features/exit/domain/repositories/exit_repository.dart';
import 'package:customer_app/features/exit/domain/usecases/poll_exit_status_use_case.dart';
import 'package:customer_app/features/exit/presentation/bloc/exit_bloc.dart';

class MockExitRepository implements ExitRepository {
  List<ExitStatusResult> pollSequence = [];
  int pollCount = 0;
  String? denyReasonToTest;

  @override
  Future<ExitStatus> pollExitStatus(String orderId) async {
    pollCount++;
    if (pollSequence.isNotEmpty) {
      final next = pollSequence.removeAt(0);
      return ExitStatus(result: next);
    }
    if (denyReasonToTest != null) {
      if (denyReasonToTest == 'QR_EXPIRED') {
        return const ExitStatus(result: ExitStatusResult.expired);
      } else {
        return const ExitStatus(result: ExitStatusResult.helpNeeded);
      }
    }
    return const ExitStatus(result: ExitStatusResult.pending);
  }
}

void main() {
  late MockExitRepository repo;
  late ExitBloc bloc;

  setUp(() {
    repo = MockExitRepository();
    bloc = ExitBloc(pollExitStatusUseCase: PollExitStatusUseCase(repo));
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is ExitInitial', () {
    expect(bloc.state, isA<ExitInitial>());
  });

  blocTest<ExitBloc, ExitState>(
    'poll sequence PENDING -> AWAITING_RFID -> OPENED transitions correctly and stops polling after OPENED',
    build: () {
      repo.pollSequence = [
        ExitStatusResult.pending,
        ExitStatusResult.awaitingRfid,
        ExitStatusResult.opened,
      ];
      return bloc;
    },
    act: (bloc) async {
      bloc.add(ExitScreenOpened(
        orderId: 'ord-100',
        token: 'token-abc',
        expiresAt: DateTime.now().add(const Duration(minutes: 10)),
      ));
      // Trigger poll ticks manually
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(ExitStatusPollTicked());
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(ExitStatusPollTicked());
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(ExitStatusPollTicked());
      await Future.delayed(const Duration(milliseconds: 10));
      // Attempt extra tick after OPENED to verify terminal state stops polling
      bloc.add(ExitStatusPollTicked());
    },
    expect: () => [
      isA<ExitDisplayingQr>(),
      isA<ExitAwaitingRfid>(),
      isA<ExitOpened>(),
    ],
    verify: (_) {
      expect(repo.pollCount, equals(3)); // No 4th poll was executed after OPENED!
    },
  );

  blocTest<ExitBloc, ExitState>(
    'QR_EXPIRED from server -> ExitTokenExpired state',
    build: () {
      repo.denyReasonToTest = 'QR_EXPIRED';
      return bloc;
    },
    act: (bloc) async {
      bloc.add(ExitScreenOpened(
        orderId: 'ord-100',
        token: 'token-abc',
        expiresAt: DateTime.now().add(const Duration(minutes: 10)),
      ));
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(ExitStatusPollTicked());
    },
    expect: () => [
      isA<ExitDisplayingQr>(),
      isA<ExitTokenExpired>(),
    ],
  );

  blocTest<ExitBloc, ExitState>(
    'Any other DENY reason -> ExitHelpNeeded state (without leaking raw reason)',
    build: () {
      repo.denyReasonToTest = 'WRONG_STORE';
      return bloc;
    },
    act: (bloc) async {
      bloc.add(ExitScreenOpened(
        orderId: 'ord-100',
        token: 'token-abc',
        expiresAt: DateTime.now().add(const Duration(minutes: 10)),
      ));
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(ExitStatusPollTicked());
    },
    expect: () => [
      isA<ExitDisplayingQr>(),
      isA<ExitHelpNeeded>(),
    ],
  );

  blocTest<ExitBloc, ExitState>(
    'Local countdown reaching zero -> ExitTokenExpired locally, late-arriving OPENED does NOT override terminal state',
    build: () {
      repo.pollSequence = [ExitStatusResult.opened];
      return bloc;
    },
    act: (bloc) async {
      bloc.add(ExitScreenOpened(
        orderId: 'ord-100',
        token: 'token-abc',
        expiresAt: DateTime.now().subtract(const Duration(seconds: 1)),
      ));
      await Future.delayed(const Duration(milliseconds: 10));
      // Late arriving OPENED tick
      bloc.add(ExitStatusPollTicked());
    },
    expect: () => [
      isA<ExitTokenExpired>(),
    ],
  );
}

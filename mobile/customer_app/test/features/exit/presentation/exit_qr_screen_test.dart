import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/exit/domain/entities/exit_status.dart';
import 'package:customer_app/features/exit/domain/repositories/exit_repository.dart';
import 'package:customer_app/features/exit/domain/usecases/poll_exit_status_use_case.dart';
import 'package:customer_app/features/exit/presentation/bloc/exit_bloc.dart';
import 'package:customer_app/features/exit/presentation/screens/exit_qr_screen.dart';

class MockExitRepoForScreen implements ExitRepository {
  @override
  Future<ExitStatus> pollExitStatus(String orderId) async {
    return const ExitStatus(result: ExitStatusResult.pending);
  }
}

void main() {
  testWidgets('Countdown color changes to warning tone under 2 minutes remaining (< 120s)', (tester) async {
    final repo = MockExitRepoForScreen();
    final bloc = ExitBloc(pollExitStatusUseCase: PollExitStatusUseCase(repo));

    // Expires in 90 seconds (< 120s)
    final expiresAt = DateTime.now().add(const Duration(seconds: 90));

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<ExitBloc>.value(
          value: bloc,
          child: ExitQrScreen(
            orderId: 'ord-warn-123',
            token: 'sample-exit-token',
            expiresAt: expiresAt,
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    final Text timerWidget = tester.widget(find.byKey(const Key('exit_timer_text')));
    expect(timerWidget.style?.color, equals(Colors.orange[800]));

    bloc.close();
  });
}

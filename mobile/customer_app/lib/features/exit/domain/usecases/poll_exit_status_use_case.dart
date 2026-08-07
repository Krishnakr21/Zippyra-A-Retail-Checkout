import '../entities/exit_status.dart';
import '../repositories/exit_repository.dart';

class PollExitStatusUseCase {
  final ExitRepository repository;

  PollExitStatusUseCase(this.repository);

  Future<ExitStatus> call(String orderId) {
    return repository.pollExitStatus(orderId);
  }
}

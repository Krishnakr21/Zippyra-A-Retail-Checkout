import '../entities/exit_status.dart';

abstract class ExitRepository {
  Future<ExitStatus> pollExitStatus(String orderId);
}

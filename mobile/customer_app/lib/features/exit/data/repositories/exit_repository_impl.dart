import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/exit_status.dart';
import '../../domain/repositories/exit_repository.dart';
import '../datasources/exit_remote_data_source.dart';

class ExitRepositoryImpl implements ExitRepository {
  final ExitRemoteDataSource remoteDataSource;

  ExitRepositoryImpl({required this.remoteDataSource});

  @override
  Future<ExitStatus> pollExitStatus(String orderId) async {
    try {
      final model = await remoteDataSource.getExitStatus(orderId);

      ExitStatusResult mappedResult;
      switch (model.result) {
        case 'AWAITING_RFID':
          mappedResult = ExitStatusResult.awaitingRfid;
          break;
        case 'OPENED':
        case 'OPEN':
        case 'RFID_CONFIRMED':
        case 'STAFF_OVERRIDE':
          mappedResult = ExitStatusResult.opened;
          break;
        case 'DENY':
          if (model.reason == 'QR_EXPIRED') {
            mappedResult = ExitStatusResult.expired;
          } else {
            mappedResult = ExitStatusResult.helpNeeded;
          }
          break;
        case 'PENDING':
        default:
          mappedResult = ExitStatusResult.pending;
          break;
      }

      return ExitStatus(result: mappedResult, gateId: model.gateId);
    } catch (e) {
      // Any network or unexpected error during exit polling safe-fails to pending/helpNeeded
      return const ExitStatus(result: ExitStatusResult.pending);
    }
  }
}

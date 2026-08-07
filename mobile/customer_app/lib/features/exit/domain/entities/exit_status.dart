import 'package:equatable/equatable.dart';

enum ExitStatusResult {
  pending,
  awaitingRfid,
  opened,
  expired,
  helpNeeded,
}

class ExitStatus extends Equatable {
  final ExitStatusResult result;
  final String? gateId;

  const ExitStatus({
    required this.result,
    this.gateId,
  });

  @override
  List<Object?> get props => [result, gateId];
}

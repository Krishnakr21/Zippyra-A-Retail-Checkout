import 'package:equatable/equatable.dart';

class StaffShiftEntity extends Equatable {
  final String id;
  final String staffId;
  final String storeId;
  final DateTime startedAt;
  final DateTime? endedAt;

  const StaffShiftEntity({
    required this.id,
    required this.staffId,
    required this.storeId,
    required this.startedAt,
    this.endedAt,
  });

  @override
  List<Object?> get props => [id, staffId, storeId, startedAt, endedAt];
}

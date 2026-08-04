import '../entities/staff_shift.dart';

abstract class ShiftRepository {
  Future<StaffShiftEntity> startShift();
  Future<void> endShift();
  Future<StaffShiftEntity?> getCurrentShift();
}

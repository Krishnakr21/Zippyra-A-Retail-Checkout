import '../../domain/entities/staff_shift.dart';
import '../../domain/repositories/shift_repository.dart';
import '../datasources/shift_remote_data_source.dart';

class ShiftRepositoryImpl implements ShiftRepository {
  final ShiftRemoteDataSource remoteDataSource;

  ShiftRepositoryImpl({required this.remoteDataSource});

  @override
  Future<StaffShiftEntity> startShift() async {
    return await remoteDataSource.startShift();
  }

  @override
  Future<void> endShift() async {
    await remoteDataSource.endShift();
  }

  @override
  Future<StaffShiftEntity?> getCurrentShift() async {
    return await remoteDataSource.getCurrentShift();
  }
}

import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/features/shift/domain/entities/staff_shift.dart';
import 'package:staff_app/features/shift/domain/repositories/shift_repository.dart';
import 'package:staff_app/features/shift/presentation/bloc/shift_bloc.dart';

class MockShiftRepository implements ShiftRepository {
  StaffShiftEntity? startShiftResult;

  @override
  Future<StaffShiftEntity> startShift() async {
    if (startShiftResult != null) {
      return startShiftResult!;
    }
    return StaffShiftEntity(
      id: 'shift-1',
      staffId: 'staff-1',
      storeId: 'store-1',
      startedAt: DateTime.now(),
    );
  }

  @override
  Future<void> endShift() async {}

  @override
  Future<StaffShiftEntity?> getCurrentShift() async => startShiftResult;
}

void main() {
  late MockShiftRepository mockShiftRepository;
  late ShiftBloc shiftBloc;

  final startedTime = DateTime.parse('2026-08-01T00:00:00Z');
  final existingShift = StaffShiftEntity(
    id: 'shift-101',
    staffId: 'staff-1',
    storeId: 'store-1',
    startedAt: startedTime,
  );

  setUp(() {
    mockShiftRepository = MockShiftRepository();
    shiftBloc = ShiftBloc(shiftRepository: mockShiftRepository);
  });

  tearDown(() {
    shiftBloc.close();
  });

  group('ShiftBloc SHIFT_ALREADY_ACTIVE', () {
    blocTest<ShiftBloc, ShiftState>(
      'ShiftStartRequested receiving existing shift results in active shift UI state',
      build: () {
        mockShiftRepository.startShiftResult = existingShift;
        return shiftBloc;
      },
      act: (bloc) => bloc.add(ShiftStartRequested()),
      expect: () => [
        ShiftLoading(),
        isA<ShiftActive>().having((s) => s.startTime, 'startTime', startedTime),
      ],
    );
  });
}

import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/referral_info.dart';
import '../../domain/repositories/referral_repository.dart';

abstract class ReferralState {}

class ReferralInitial extends ReferralState {}

class ReferralLoading extends ReferralState {}

class ReferralLoaded extends ReferralState {
  final ReferralInfo info;
  ReferralLoaded(this.info);
}

class ReferralError extends ReferralState {
  final String message;
  ReferralError(this.message);
}

class ReferralApplySuccess extends ReferralState {}

class ReferralCubit extends Cubit<ReferralState> {
  final ReferralRepository repository;

  ReferralCubit({required this.repository}) : super(ReferralInitial());

  Future<void> loadReferralInfo() async {
    emit(ReferralLoading());
    try {
      final info = await repository.getReferralInfo();
      emit(ReferralLoaded(info));
    } catch (e) {
      emit(ReferralError('Failed to load referral info: $e'));
    }
  }

  Future<void> applyReferralCode(String code) async {
    emit(ReferralLoading());
    try {
      await repository.applyReferralCode(code);
      emit(ReferralApplySuccess());
    } catch (e) {
      emit(ReferralError('Failed to apply referral code: $e'));
    }
  }
}

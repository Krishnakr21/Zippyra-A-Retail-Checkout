import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/loyalty_balance.dart';
import '../../domain/usecases/get_loyalty_balance_use_case.dart';

part 'loyalty_event.dart';
part 'loyalty_state.dart';

class LoyaltyBloc extends Bloc<LoyaltyEvent, LoyaltyState> {
  final GetLoyaltyBalanceUseCase getLoyaltyBalanceUseCase;

  LoyaltyBloc({required this.getLoyaltyBalanceUseCase}) : super(LoyaltyInitial()) {
    on<LoyaltyBalanceRequested>(_onLoyaltyBalanceRequested);
  }

  Future<void> _onLoyaltyBalanceRequested(
    LoyaltyBalanceRequested event,
    Emitter<LoyaltyState> emit,
  ) async {
    if (!event.refresh && state is LoyaltyLoaded) {
      return;
    }

    if (state is! LoyaltyLoaded) {
      emit(LoyaltyLoading());
    }

    try {
      final balance = await getLoyaltyBalanceUseCase();
      emit(LoyaltyLoaded(
        pointsBalance: balance.pointsBalance,
        pointsReserved: balance.pointsReserved,
        tier: balance.tier,
        tierDisplayName: balance.tierDisplayName,
        lifetimePointsEarned: balance.lifetimePointsEarned,
        pointsToNextTier: balance.pointsToNextTier,
        nextTierName: balance.nextTierName,
      ));
    } catch (e) {
      emit(LoyaltyError(e.toString()));
    }
  }
}

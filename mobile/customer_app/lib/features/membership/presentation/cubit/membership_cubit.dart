import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/subscription_plan.dart';
import '../../domain/entities/member_subscription.dart';
import '../../domain/repositories/membership_repository.dart';

abstract class MembershipState {}

class MembershipInitial extends MembershipState {}

class MembershipLoading extends MembershipState {}

class MembershipLoaded extends MembershipState {
  final List<SubscriptionPlan> plans;
  final MemberSubscription? activeSubscription;

  MembershipLoaded({
    required this.plans,
    this.activeSubscription,
  });
}

class MembershipError extends MembershipState {
  final String message;
  MembershipError(this.message);
}

class MembershipCubit extends Cubit<MembershipState> {
  final MembershipRepository repository;

  MembershipCubit({required this.repository}) : super(MembershipInitial());

  Future<void> loadMembershipData() async {
    emit(MembershipLoading());
    try {
      final plans = await repository.getPlans();
      final activeSub = await repository.getMySubscription();
      emit(MembershipLoaded(plans: plans, activeSubscription: activeSub));
    } catch (e) {
      emit(MembershipError('Failed to load membership details: $e'));
    }
  }

  Future<void> subscribe(String planId) async {
    emit(MembershipLoading());
    try {
      await repository.subscribe(planId);
      await loadMembershipData();
    } catch (e) {
      emit(MembershipError('Failed to subscribe: $e'));
    }
  }

  Future<void> cancelSubscription() async {
    emit(MembershipLoading());
    try {
      await repository.cancelSubscription();
      await loadMembershipData();
    } catch (e) {
      emit(MembershipError('Failed to cancel subscription: $e'));
    }
  }
}

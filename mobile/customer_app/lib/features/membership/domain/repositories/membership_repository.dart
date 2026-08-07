import '../entities/subscription_plan.dart';
import '../entities/member_subscription.dart';

abstract class MembershipRepository {
  Future<List<SubscriptionPlan>> getPlans();
  Future<MemberSubscription?> getMySubscription();
  Future<MemberSubscription> subscribe(String planId);
  Future<void> cancelSubscription();
}

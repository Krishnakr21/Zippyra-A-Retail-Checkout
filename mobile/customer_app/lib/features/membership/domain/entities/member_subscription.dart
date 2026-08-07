import 'subscription_plan.dart';

class MemberSubscription {
  final String id;
  final String userId;
  final String planId;
  final String status; // ACTIVE, PENDING, CANCELLED, EXPIRED
  final String? razorpaySubscriptionId;
  final DateTime? currentPeriodEnd;
  final DateTime createdAt;
  final SubscriptionPlan? plan;

  const MemberSubscription({
    required this.id,
    required this.userId,
    required this.planId,
    required this.status,
    this.razorpaySubscriptionId,
    this.currentPeriodEnd,
    required this.createdAt,
    this.plan,
  });

  bool get isActive => status == 'ACTIVE';
}

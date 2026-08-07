import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/subscription_plan.dart';
import '../../domain/entities/member_subscription.dart';

abstract class MembershipRemoteDataSource {
  Future<List<SubscriptionPlan>> getPlans();
  Future<MemberSubscription?> getMySubscription();
  Future<MemberSubscription> subscribe(String planId);
  Future<void> cancelSubscription();
}

class MembershipRemoteDataSourceImpl implements MembershipRemoteDataSource {
  final ApiClient apiClient;

  MembershipRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<SubscriptionPlan>> getPlans() async {
    try {
      final response = await apiClient.get('/v1/subscription/plans?chain_id=chain-hq-001');
      final data = response.data as Map<String, dynamic>;
      final list = (data['plans'] as List<dynamic>? ?? []);
      return list.map((json) {
        final map = json as Map<String, dynamic>;
        final benefits = map['benefits'] as Map<String, dynamic>? ?? {};
        return SubscriptionPlan(
          id: map['id'] as String? ?? '',
          chainId: map['chain_id'] as String? ?? '',
          name: map['name'] as String? ?? '',
          pricePaise: (map['price_paise'] as num? ?? 0).toInt(),
          billingInterval: map['billing_interval'] as String? ?? 'MONTHLY',
          loyaltyMultiplierBonus: (benefits['loyalty_multiplier_bonus'] as num? ?? 0.5).toDouble(),
          freeDelivery: benefits['free_delivery'] as bool? ?? true,
          isActive: map['is_active'] as bool? ?? true,
        );
      }).toList();
    } catch (e) {
      throw ServerFailure('Failed to fetch subscription plans: $e');
    }
  }

  @override
  Future<MemberSubscription?> getMySubscription() async {
    try {
      final response = await apiClient.get('/v1/subscription/mine');
      final data = response.data as Map<String, dynamic>;
      final subMap = data['subscription'] as Map<String, dynamic>?;
      if (subMap == null) return null;

      SubscriptionPlan? plan;
      if (subMap['plan'] != null) {
        final pMap = subMap['plan'] as Map<String, dynamic>;
        final benefits = pMap['benefits'] as Map<String, dynamic>? ?? {};
        plan = SubscriptionPlan(
          id: pMap['id'] as String? ?? '',
          chainId: pMap['chain_id'] as String? ?? '',
          name: pMap['name'] as String? ?? '',
          pricePaise: (pMap['price_paise'] as num? ?? 0).toInt(),
          billingInterval: pMap['billing_interval'] as String? ?? 'MONTHLY',
          loyaltyMultiplierBonus: (benefits['loyalty_multiplier_bonus'] as num? ?? 0.5).toDouble(),
          freeDelivery: benefits['free_delivery'] as bool? ?? true,
          isActive: pMap['is_active'] as bool? ?? true,
        );
      }

      return MemberSubscription(
        id: subMap['id'] as String? ?? '',
        userId: subMap['user_id'] as String? ?? '',
        planId: subMap['plan_id'] as String? ?? '',
        status: subMap['status'] as String? ?? 'PENDING',
        razorpaySubscriptionId: subMap['razorpay_subscription_id'] as String?,
        currentPeriodEnd: subMap['current_period_end'] != null
            ? DateTime.tryParse(subMap['current_period_end'] as String)
            : null,
        createdAt: DateTime.tryParse(subMap['created_at'] as String? ?? '') ?? DateTime.now(),
        plan: plan,
      );
    } catch (e) {
      throw ServerFailure('Failed to fetch subscription: $e');
    }
  }

  @override
  Future<MemberSubscription> subscribe(String planId) async {
    try {
      final response = await apiClient.post('/v1/subscription/subscribe', data: {
        'plan_id': planId,
      });
      final map = response.data as Map<String, dynamic>;
      return MemberSubscription(
        id: map['subscription_id'] as String? ?? '',
        userId: '',
        planId: map['plan_id'] as String? ?? planId,
        status: map['status'] as String? ?? 'ACTIVE',
        razorpaySubscriptionId: map['razorpay_subscription_id'] as String?,
        currentPeriodEnd: map['current_period_end'] != null
            ? DateTime.tryParse(map['current_period_end'] as String)
            : null,
        createdAt: DateTime.now(),
      );
    } catch (e) {
      throw ServerFailure('Failed to subscribe to plan: $e');
    }
  }

  @override
  Future<void> cancelSubscription() async {
    try {
      await apiClient.post('/v1/subscription/cancel');
    } catch (e) {
      throw ServerFailure('Failed to cancel subscription: $e');
    }
  }
}

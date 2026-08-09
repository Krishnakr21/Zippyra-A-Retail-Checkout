import '../../domain/entities/subscription_plan.dart';
import '../../domain/entities/member_subscription.dart';
import '../../domain/repositories/membership_repository.dart';
import '../datasources/membership_remote_data_source.dart';

class MembershipRepositoryImpl implements MembershipRepository {
  final MembershipRemoteDataSource remoteDataSource;

  MembershipRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<SubscriptionPlan>> getPlans() {
    return remoteDataSource.getPlans();
  }

  @override
  Future<MemberSubscription?> getMySubscription() {
    return remoteDataSource.getMySubscription();
  }

  @override
  Future<MemberSubscription> subscribe(String planId) {
    return remoteDataSource.subscribe(planId);
  }

  @override
  Future<void> cancelSubscription() {
    return remoteDataSource.cancelSubscription();
  }
}

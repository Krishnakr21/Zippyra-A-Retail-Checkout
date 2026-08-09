import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/referral_info.dart';

abstract class ReferralRemoteDataSource {
  Future<ReferralInfo> getReferralInfo();
  Future<void> applyReferralCode(String referralCode);
}

class ReferralRemoteDataSourceImpl implements ReferralRemoteDataSource {
  final ApiClient apiClient;

  ReferralRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<ReferralInfo> getReferralInfo() async {
    try {
      final response = await apiClient.get('/v1/loyalty/referral-code');
      final data = response.data as Map<String, dynamic>;
      return ReferralInfo(
        referralCode: data['referral_code'] as String? ?? '',
        shareText: data['share_text'] as String? ?? '',
        referrerRewardPoints: (data['referrer_reward_points'] as num? ?? 100).toInt(),
        referredRewardPoints: (data['referred_reward_points'] as num? ?? 50).toInt(),
      );
    } catch (e) {
      throw ServerFailure('Failed to fetch referral info: $e');
    }
  }

  @override
  Future<void> applyReferralCode(String referralCode) async {
    try {
      await apiClient.post('/v1/loyalty/referral/apply', data: {
        'referral_code': referralCode,
      });
    } catch (e) {
      throw ServerFailure('Failed to apply referral code: $e');
    }
  }
}

import '../entities/referral_info.dart';

abstract class ReferralRepository {
  Future<ReferralInfo> getReferralInfo();
  Future<void> applyReferralCode(String referralCode);
}

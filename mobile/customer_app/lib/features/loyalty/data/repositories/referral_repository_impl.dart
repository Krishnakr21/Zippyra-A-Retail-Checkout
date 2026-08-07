import '../../domain/entities/referral_info.dart';
import '../../domain/repositories/referral_repository.dart';
import '../datasources/referral_remote_data_source.dart';

class ReferralRepositoryImpl implements ReferralRepository {
  final ReferralRemoteDataSource remoteDataSource;

  ReferralRepositoryImpl({required this.remoteDataSource});

  @override
  Future<ReferralInfo> getReferralInfo() {
    return remoteDataSource.getReferralInfo();
  }

  @override
  Future<void> applyReferralCode(String referralCode) {
    return remoteDataSource.applyReferralCode(referralCode);
  }
}

import '../../domain/models/privacy_consent.dart';
import '../../domain/models/dpdp_request.dart';
import '../../domain/models/grievance_officer.dart';
import '../../domain/repositories/privacy_repository.dart';
import '../datasources/privacy_remote_data_source.dart';

class PrivacyRepositoryImpl implements PrivacyRepository {
  final PrivacyRemoteDataSource remoteDataSource;

  PrivacyRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<PrivacyConsent>> getConsents() => remoteDataSource.getConsents();

  @override
  Future<PrivacyConsent> updateConsent(String consentType, bool granted) =>
      remoteDataSource.updateConsent(consentType, granted);

  @override
  Future<DPDPRequest> submitRequest(String requestType, {String? detail}) =>
      remoteDataSource.submitRequest(requestType, detail: detail);

  @override
  Future<List<DPDPRequest>> getMyRequests() => remoteDataSource.getMyRequests();

  @override
  Future<GrievanceOfficer> getGrievanceOfficer() => remoteDataSource.getGrievanceOfficer();
}

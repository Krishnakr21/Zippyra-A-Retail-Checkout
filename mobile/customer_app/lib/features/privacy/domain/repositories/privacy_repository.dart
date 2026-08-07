import '../models/privacy_consent.dart';
import '../models/dpdp_request.dart';
import '../models/grievance_officer.dart';

abstract class PrivacyRepository {
  Future<List<PrivacyConsent>> getConsents();
  Future<PrivacyConsent> updateConsent(String consentType, bool granted);
  Future<DPDPRequest> submitRequest(String requestType, {String? detail});
  Future<List<DPDPRequest>> getMyRequests();
  Future<GrievanceOfficer> getGrievanceOfficer();
}

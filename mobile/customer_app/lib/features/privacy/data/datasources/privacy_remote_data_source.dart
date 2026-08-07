import '../../domain/models/privacy_consent.dart';
import '../../domain/models/dpdp_request.dart';
import '../../domain/models/grievance_officer.dart';

abstract class PrivacyRemoteDataSource {
  Future<List<PrivacyConsent>> getConsents();
  Future<PrivacyConsent> updateConsent(String consentType, bool granted);
  Future<DPDPRequest> submitRequest(String requestType, {String? detail});
  Future<List<DPDPRequest>> getMyRequests();
  Future<GrievanceOfficer> getGrievanceOfficer();
}

class MockPrivacyRemoteDataSource implements PrivacyRemoteDataSource {
  final Map<String, PrivacyConsent> _consents = {
    'MARKETING_COMMS': const PrivacyConsent(consentType: 'MARKETING_COMMS', granted: true, consentVersion: 'v1.0'),
    'LOCATION_TRACKING': const PrivacyConsent(consentType: 'LOCATION_TRACKING', granted: false, consentVersion: 'v0.9', needsReconfirmation: true),
    'ANALYTICS_SHARING': const PrivacyConsent(consentType: 'ANALYTICS_SHARING', granted: true, consentVersion: 'v1.0'),
  };

  final List<DPDPRequest> _myRequests = [];

  @override
  Future<List<PrivacyConsent>> getConsents() async {
    return _consents.values.toList();
  }

  @override
  Future<PrivacyConsent> updateConsent(String consentType, bool granted) async {
    final updated = PrivacyConsent(
      consentType: consentType,
      granted: granted,
      consentVersion: 'v1.0',
      needsReconfirmation: false, // Badge disappears after re-toggling / updating!
    );
    _consents[consentType] = updated;
    return updated;
  }

  @override
  Future<DPDPRequest> submitRequest(String requestType, {String? detail}) async {
    final req = DPDPRequest(
      id: 'req_${DateTime.now().millisecondsSinceEpoch}',
      requestType: requestType,
      status: 'RECEIVED',
      detail: detail,
      createdAt: DateTime.now().toIso8601String(),
    );
    _myRequests.insert(0, req);
    return req;
  }

  @override
  Future<List<DPDPRequest>> getMyRequests() async {
    return List.unmodifiable(_myRequests);
  }

  @override
  Future<GrievanceOfficer> getGrievanceOfficer() async {
    return const GrievanceOfficer(
      name: 'Nisha Sharma',
      title: 'Data Protection & Grievance Officer',
      email: 'grievance@zippyra.com',
      address: 'Zippyra India Tech Pvt Ltd, 4th Floor, HSR Layout, Bengaluru 560102',
      acknowledgmentSla: '72 hours',
    );
  }
}

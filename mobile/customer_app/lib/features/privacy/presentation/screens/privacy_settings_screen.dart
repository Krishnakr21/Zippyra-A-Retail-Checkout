import 'package:flutter/material.dart';
import '../../domain/models/privacy_consent.dart';
import '../../domain/models/dpdp_request.dart';
import '../../domain/models/grievance_officer.dart';
import '../../domain/repositories/privacy_repository.dart';

class PrivacySettingsScreen extends StatefulWidget {
  final PrivacyRepository repository;

  const PrivacySettingsScreen({Key? key, required this.repository}) : super(key: key);

  @override
  State<PrivacySettingsScreen> createState() => _PrivacySettingsScreenState();
}

class _PrivacySettingsScreenState extends State<PrivacySettingsScreen> {
  bool _isLoading = true;
  String? _errorMessage;

  List<PrivacyConsent> _consents = [];
  List<DPDPRequest> _myRequests = [];
  GrievanceOfficer? _grievanceOfficer;

  @override
  void initState() {
    super.initState();
    _loadPrivacyData();
  }

  Future<void> _loadPrivacyData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final consents = await widget.repository.getConsents();
      final requests = await widget.repository.getMyRequests();
      final officer = await widget.repository.getGrievanceOfficer();

      setState(() {
        _consents = consents;
        _myRequests = requests;
        _grievanceOfficer = officer;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _toggleConsent(PrivacyConsent consent, bool newValue) async {
    // Optimistic UI update
    final index = _consents.indexWhere((c) => c.consentType == consent.consentType);
    if (index == -1) return;

    final original = _consents[index];
    setState(() {
      _consents[index] = consent.copyWith(
        granted: newValue,
        needsReconfirmation: false, // Badge disappears after re-toggling
      );
    });

    try {
      final updated = await widget.repository.updateConsent(consent.consentType, newValue);
      setState(() {
        _consents[index] = updated;
      });
    } catch (e) {
      // Rollback on error
      setState(() {
        _consents[index] = original;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to update consent: $e')),
        );
      }
    }
  }

  Future<void> _handleRequestData() async {
    try {
      final req = await widget.repository.submitRequest('ACCESS', detail: 'User requested copy of personal data');
      setState(() {
        _myRequests.insert(0, req);
      });

      if (!mounted) return;
      showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Data Access Request Submitted'),
          content: const Text("We'll email/notify you within 30 days with your data export package."),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('OK'),
            ),
          ],
        ),
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to submit request: $e')),
        );
      }
    }
  }

  Future<void> _handleDeleteAccount() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Account Deletion Request'),
        content: const Text(
          'Submitting this request will initiate an administrative review to permanently delete your personal data. '
          'Statutory invoice records will be retained as required by GST law. Are you sure you want to proceed?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Confirm Request', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      final req = await widget.repository.submitRequest('DELETION', detail: 'User requested account & PII deletion');
      setState(() {
        _myRequests.insert(0, req);
      });

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Deletion request submitted for administrative review')),
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to submit deletion request: $e')),
        );
      }
    }
  }

  String _getConsentTitle(String consentType) {
    switch (consentType) {
      case 'MARKETING_COMMS':
        return 'Marketing Communications';
      case 'LOCATION_TRACKING':
        return 'Location Personalization';
      case 'ANALYTICS_SHARING':
        return 'Analytics Sharing';
      default:
        return consentType;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Privacy & Data Protection'),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(child: Text('Error: $_errorMessage'))
              : ListView(
                  padding: const EdgeInsets.all(16.0),
                  children: [
                    // Section 1: Consents & Preferences
                    Text(
                      'Consent Preferences',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),

                    ..._consents.map((consent) {
                      return Card(
                        key: Key('consent_card_${consent.consentType}'),
                        margin: const EdgeInsets.symmetric(vertical: 4.0),
                        child: SwitchListTile(
                          title: Row(
                            children: [
                              Expanded(
                                child: Text(_getConsentTitle(consent.consentType)),
                              ),
                              if (consent.needsReconfirmation) ...[
                                const SizedBox(width: 8),
                                Container(
                                  key: Key('reconfirmation_badge_${consent.consentType}'),
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: Colors.amber.shade700,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: const Text(
                                    'Re-confirmation Required',
                                    style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                                  ),
                                ),
                              ],
                            ],
                          ),
                          value: consent.granted,
                          onChanged: (val) => _toggleConsent(consent, val),
                        ),
                      );
                    }).toList(),

                    const SizedBox(height: 24),

                    // Section 2: Data Rights Actions
                    Text(
                      'Your Data Rights (DPDP Act 2023)',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),

                    ElevatedButton.icon(
                      key: const Key('request_data_btn'),
                      onPressed: _handleRequestData,
                      icon: const Icon(Icons.download),
                      label: const Text('Request My Data'),
                    ),
                    const SizedBox(height: 8),

                    OutlinedButton.icon(
                      key: const Key('delete_account_btn'),
                      style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
                      onPressed: _handleDeleteAccount,
                      icon: const Icon(Icons.delete_forever),
                      label: const Text('Delete My Account & Data'),
                    ),

                    const SizedBox(height: 24),

                    // Section 3: Request History
                    Text(
                      'Request History',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),

                    if (_myRequests.isEmpty)
                      const Text('No privacy requests submitted yet.', style: TextStyle(color: Colors.grey))
                    else
                      ..._myRequests.map((req) {
                        final isAccessCompleted = req.requestType == 'ACCESS' && req.status == 'COMPLETED';
                        return Card(
                          child: ListTile(
                            title: Text('${req.requestType} Request'),
                            subtitle: Text(
                              isAccessCompleted
                                  ? 'Status: Export Assembled & Ready (Valid for 7 days)'
                                  : 'Status: ${req.status}',
                            ),
                            trailing: isAccessCompleted
                                ? ElevatedButton.icon(
                                    key: Key('download_export_btn_${req.id}'),
                                    style: ElevatedButton.styleFrom(backgroundColor: Colors.green),
                                    onPressed: () {
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        const SnackBar(content: Text('Downloading presigned DPDP data export...')),
                                      );
                                    },
                                    icon: const Icon(Icons.file_download, size: 18),
                                    label: const Text('Download Export'),
                                  )
                                : Text(
                                    req.status == 'COMPLETED' ? 'Completed' : 'Pending Review',
                                    style: TextStyle(
                                      color: req.status == 'COMPLETED' ? Colors.green : Colors.orange,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                          ),
                        );
                      }).toList(),

                    const SizedBox(height: 24),

                    // Section 4: Grievance Officer Contact
                    if (_grievanceOfficer != null) ...[
                      Text(
                        'Grievance Officer',
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
                      ),
                      const SizedBox(height: 8),
                      Card(
                        color: Colors.blueGrey.shade50,
                        child: Padding(
                          padding: const EdgeInsets.all(12.0),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _grievanceOfficer!.name,
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Colors.black87),
                              ),
                              Text(_grievanceOfficer!.title, style: const TextStyle(color: Colors.black54)),
                              const SizedBox(height: 4),
                              SelectableText('Email: ${_grievanceOfficer!.email}', style: const TextStyle(color: Colors.blue)),
                              const SizedBox(height: 4),
                              Text(_grievanceOfficer!.address, style: const TextStyle(fontSize: 12, color: Colors.black54)),
                              const SizedBox(height: 8),
                              Text(
                                'Statutory Acknowledgment Commitment: Within ${_grievanceOfficer!.acknowledgmentSla}',
                                style: const TextStyle(fontSize: 11, fontStyle: FontStyle.italic, color: Colors.black87),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
    );
  }
}

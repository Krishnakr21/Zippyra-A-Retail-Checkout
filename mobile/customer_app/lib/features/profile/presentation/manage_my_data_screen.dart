import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class ManageMyDataScreen extends StatelessWidget {
  const ManageMyDataScreen({super.key});

  void _showDownloadDataModal(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.file_download_rounded, color: Color(0xFF2563EB), size: 48),
              const SizedBox(height: 12),
              const Text(
                'Request Data Export',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B)),
              ),
              const SizedBox(height: 8),
              const Text(
                'We will bundle your account profile, order receipts, and scan history into an encrypted JSON/CSV archive and email it to your registered email.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Color(0xFF64748B), fontSize: 13, height: 1.4),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF16A34A),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  onPressed: () {
                    Navigator.pop(ctx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text('Data export request submitted! Check your email shortly. ✓'),
                        backgroundColor: Color(0xFF16A34A),
                      ),
                    );
                  },
                  child: const Text('Send My Data Archive', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showDeleteDataConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text('Delete All My Data?', style: TextStyle(fontWeight: FontWeight.bold)),
        content: const Text(
          'This permanent action will wipe all your personal information, loyalty points balance, saved addresses, and scan history from Zippyra servers in accordance with DPDPA 2023 regulations.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogCtx),
            child: const Text('Cancel', style: TextStyle(color: Colors.grey)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFFDC2626)),
            onPressed: () {
              Navigator.pop(dialogCtx);
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text('Data erasure request submitted under DPDP guidelines.'),
                  backgroundColor: Color(0xFFDC2626),
                ),
              );
            },
            child: const Text('Delete Permanently', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        iconTheme: const IconThemeData(color: Color(0xFF1E293B)),
        title: Row(
          children: const [
            Text(
              'Manage My Data',
              style: TextStyle(fontWeight: FontWeight.w900, fontSize: 18, color: Color(0xFF1E293B)),
            ),
            SizedBox(width: 6),
            Text('🛡️', style: TextStyle(fontSize: 16)),
          ],
        ),
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Top Security Info Banner (Figma Screen A-15)
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFEFF6FF),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xFFBFDBFE)),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: const [
                  Text('🔒', style: TextStyle(fontSize: 22)),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      'Your data is stored encrypted in India-only servers. Zippyra never sells personal data to advertisers.',
                      style: TextStyle(color: Color(0xFF1E3A8A), fontSize: 12, fontWeight: FontWeight.w600, height: 1.4),
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 20),

            // Options Card 1 (Figma Screen A-15)
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2E8F0)),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              child: Column(
                children: [
                  _buildOptionTile(
                    icon: '📋',
                    title: 'Download My Data',
                    subtitle: 'GDPR / DPDPA compliant archive',
                    onTap: () => _showDownloadDataModal(context),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildOptionTile(
                    icon: '✉️',
                    title: 'Communication Preferences',
                    subtitle: 'Email, SMS, WhatsApp notifications',
                    onTap: () => context.push('/notifications'),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildOptionTile(
                    icon: '🗑️',
                    title: 'Delete All My Data',
                    subtitle: 'Permanent action',
                    textColor: const Color(0xFFDC2626),
                    onTap: () => _showDeleteDataConfirmation(context),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 20),

            // Options Card 2 (Figma Screen A-15)
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2E8F0)),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              child: Column(
                children: [
                  _buildOptionTile(
                    icon: '👁️',
                    title: 'Data We Collect',
                    subtitle: 'Personal info, scan history, transactions',
                    onTap: () {
                      showDialog(
                        context: context,
                        builder: (ctx) => AlertDialog(
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                          title: const Text('Data We Collect'),
                          content: const Text(
                            'Zippyra collects mobile numbers, transaction records, store check-in locations, and scanned barcode IDs solely for processing in-store checkout and receipts.',
                          ),
                          actions: [
                            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close')),
                          ],
                        ),
                      );
                    },
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildOptionTile(
                    icon: '🔗',
                    title: 'Third-Party Sharing',
                    subtitle: 'Encrypted payment gateways & store APIs',
                    onTap: () {
                      showDialog(
                        context: context,
                        builder: (ctx) => AlertDialog(
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                          title: const Text('Third-Party Sharing'),
                          content: const Text(
                            'Data is shared strictly with RBI-compliant payment aggregators (Razorpay/UPI) and store inventory engines for real-time bill generation.',
                          ),
                          actions: [
                            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close')),
                          ],
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildOptionTile({
    required String icon,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
    Color textColor = const Color(0xFF1E293B),
  }) {
    return ListTile(
      onTap: onTap,
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: const Color(0xFFF8FAFC),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(icon, style: const TextStyle(fontSize: 20)),
      ),
      title: Text(title, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: textColor)),
      subtitle: Text(subtitle, style: const TextStyle(fontSize: 11, color: Color(0xFF64748B))),
      trailing: const Icon(Icons.chevron_right, color: Color(0xFF94A3B8), size: 20),
    );
  }
}

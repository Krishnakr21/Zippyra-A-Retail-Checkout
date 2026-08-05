import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';

class StaffProfileScreen extends StatelessWidget {
  const StaffProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Staff Profile & Credentials')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const ZCard(
            child: ListTile(
              leading: CircleAvatar(backgroundColor: ZippyraColors.primaryBlue, child: Icon(Icons.badge, color: Colors.white)),
              title: Text('Staff ID #ST-8819', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: Text('Role: Cashier / Store Associate'),
            ),
          ),
          const SizedBox(height: 16),
          ZCard(
            child: ExpansionTile(
              leading: const Icon(Icons.privacy_tip_outlined, color: ZippyraColors.primaryBlue),
              title: const Text('Privacy & DPDP Rights', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: const Text('Grievance Officer & Data Requests'),
              children: [
                Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Data Protection & Grievance Officer', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                      const SizedBox(height: 4),
                      const Text('Nisha Sharma\ngrievance@zippyra.com\nAcknowledgment SLA: 72 hours', style: TextStyle(fontSize: 12, color: Colors.black54)),
                      const Divider(height: 24),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(content: Text('DPDP Access Request submitted. Summary will be emailed to staff contact.')),
                                );
                              },
                              icon: const Icon(Icons.download, size: 16),
                              label: const Text('Request My Data', style: TextStyle(fontSize: 11)),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(content: Text('DPDP Deletion Request submitted. Manager verification required.')),
                                );
                              },
                              style: OutlinedButton.styleFrom(foregroundColor: ZippyraColors.errorRed),
                              icon: const Icon(Icons.delete_forever, size: 16),
                              label: const Text('Delete Account', style: TextStyle(fontSize: 11)),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          ZCard(
            onTap: () => context.go('/auth'),
            child: const ListTile(
              leading: Icon(Icons.logout, color: ZippyraColors.errorRed),
              title: Text('End Shift & Logout', style: TextStyle(color: ZippyraColors.errorRed, fontWeight: FontWeight.bold)),
            ),
          ),
        ],
      ),
    );
  }
}

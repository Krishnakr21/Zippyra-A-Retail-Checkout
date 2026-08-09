import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  bool _locationToggle = true;
  bool _notificationsToggle = true;
  bool _whatsappToggle = true;
  bool _darkModeToggle = false;
  bool _biometricToggle = true;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        iconTheme: const IconThemeData(color: Color(0xFF1E293B)),
        title: const Text(
          'Settings',
          style: TextStyle(fontWeight: FontWeight.w900, fontSize: 18, color: Color(0xFF1E293B)),
        ),
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // 1. PREFERENCES (Figma Screen A-11)
            const Text('PREFERENCES', style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: Color(0xFF64748B), letterSpacing: 0.5)),
            const SizedBox(height: 8),
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2E8F0)),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              child: Column(
                children: [
                  _buildToggleTile(
                    icon: '📍',
                    title: 'Location',
                    subtitle: 'Auto store detect',
                    value: _locationToggle,
                    onChanged: (val) => setState(() => _locationToggle = val),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildToggleTile(
                    icon: '🔔',
                    title: 'Notifications',
                    subtitle: 'Offers, gates, orders',
                    value: _notificationsToggle,
                    onChanged: (val) => setState(() => _notificationsToggle = val),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildToggleTile(
                    icon: '💬',
                    title: 'WhatsApp Receipts',
                    subtitle: 'Auto-send after checkout',
                    value: _whatsappToggle,
                    onChanged: (val) => setState(() => _whatsappToggle = val),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  _buildToggleTile(
                    icon: '🌙',
                    title: 'Dark Mode',
                    subtitle: 'System default',
                    value: _darkModeToggle,
                    onChanged: (val) => setState(() => _darkModeToggle = val),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 20),

            // 2. SECURITY (Figma Screen A-11)
            const Text('SECURITY', style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: Color(0xFF64748B), letterSpacing: 0.5)),
            const SizedBox(height: 8),
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2E8F0)),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              child: Column(
                children: [
                  _buildToggleTile(
                    icon: '👆',
                    title: 'Biometric Lock',
                    subtitle: 'Face ID / Fingerprint unlock',
                    value: _biometricToggle,
                    onChanged: (val) => setState(() => _biometricToggle = val),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  ListTile(
                    leading: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(color: const Color(0xFFF8FAFC), borderRadius: BorderRadius.circular(12)),
                      child: const Text('🔐', style: TextStyle(fontSize: 20)),
                    ),
                    title: const Text('Payment PIN', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E293B))),
                    subtitle: const Text('Required above ₹2000', style: TextStyle(fontSize: 11, color: Color(0xFF64748B))),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFF94A3B8), size: 20),
                    onTap: () {},
                  ),
                ],
              ),
            ),

            const SizedBox(height: 20),

            // 3. DATA & PRIVACY (Figma Screen A-11)
            const Text('DATA & PRIVACY', style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: Color(0xFF64748B), letterSpacing: 0.5)),
            const SizedBox(height: 8),
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2E8F0)),
                boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              child: Column(
                children: [
                  ListTile(
                    leading: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(color: const Color(0xFFF8FAFC), borderRadius: BorderRadius.circular(12)),
                      child: const Text('🛡️', style: TextStyle(fontSize: 20)),
                    ),
                    title: const Text('Manage My Data', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E293B))),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFF94A3B8), size: 20),
                    onTap: () => context.push('/profile/data'),
                  ),
                  const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                  ListTile(
                    leading: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(color: const Color(0xFFFEF2F2), borderRadius: BorderRadius.circular(12)),
                      child: const Text('🗑️', style: TextStyle(fontSize: 20)),
                    ),
                    title: const Text('Delete Account', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFFDC2626))),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFFDC2626), size: 20),
                    onTap: () => context.push('/profile/data'),
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

  Widget _buildToggleTile({
    required String icon,
    required String title,
    required String subtitle,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(color: const Color(0xFFF8FAFC), borderRadius: BorderRadius.circular(12)),
        child: Text(icon, style: const TextStyle(fontSize: 20)),
      ),
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E293B))),
      subtitle: Text(subtitle, style: const TextStyle(fontSize: 11, color: Color(0xFF64748B))),
      trailing: Switch(
        value: value,
        activeColor: const Color(0xFF2563EB),
        onChanged: onChanged,
      ),
    );
  }
}

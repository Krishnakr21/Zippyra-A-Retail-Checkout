import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../injection_container.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  String _displayName = 'Rahul Mehta';
  String _phone = '+91 98765 43210';
  String _email = 'krishna@gmail.com';
  String _birthday = 'Mar 12, 1994';
  String _language = 'English';

  @override
  void initState() {
    super.initState();
    _loadProfileData();
  }

  Future<void> _loadProfileData() async {
    final storage = sl<SecureStorage>();
    final savedName = await storage.read(key: 'user_name');
    final savedPhone = await storage.read(key: 'user_phone');
    final savedEmail = await storage.read(key: 'user_email');
    final savedBirthday = await storage.read(key: 'user_birthday');
    final savedLang = await storage.read(key: 'user_language');

    if (mounted) {
      setState(() {
        if (savedName != null && savedName.isNotEmpty) _displayName = savedName;
        if (savedPhone != null && savedPhone.isNotEmpty) _phone = savedPhone;
        if (savedEmail != null && savedEmail.isNotEmpty) _email = savedEmail;
        if (savedBirthday != null && savedBirthday.isNotEmpty) _birthday = savedBirthday;
        if (savedLang != null && savedLang.isNotEmpty) _language = savedLang;
      });
    }
  }

  // Figma Screen A-13 LOGOUT CONFIRM Bottom Sheet
  void _showLogoutConfirmation() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (bottomSheetCtx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: const Color(0xFFE2E8F0),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 20),
              const Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Log Out?',
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.w900, color: Color(0xFF0F172A)),
                ),
              ),
              const SizedBox(height: 8),
              const Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  "You'll need to verify your mobile number again to log back in.",
                  style: TextStyle(fontSize: 13, color: Color(0xFF64748B), height: 1.4),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFDC2626),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  onPressed: () async {
                    Navigator.pop(bottomSheetCtx);
                    final storage = sl<SecureStorage>();
                    await storage.delete(key: 'access_token');
                    await storage.delete(key: 'refresh_token');
                    await storage.delete(key: 'store_session_id');
                    if (mounted) {
                      context.go('/auth');
                    }
                  },
                  child: const Text('Yes, Log Out', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 15)),
                ),
              ),
              const SizedBox(height: 10),
              SizedBox(
                width: double.infinity,
                child: TextButton(
                  style: TextButton.styleFrom(
                    backgroundColor: const Color(0xFFEFF6FF),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  onPressed: () => Navigator.pop(bottomSheetCtx),
                  child: const Text('Cancel', style: TextStyle(color: Color(0xFF1E3A8A), fontWeight: FontWeight.w800, fontSize: 15)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // Figma Screen A-14 APP RATING Bottom Sheet with Automatic Platform Detection
  void _showAppRatingModal() {
    final isAppleDevice = defaultTargetPlatform == TargetPlatform.iOS || defaultTargetPlatform == TargetPlatform.macOS;
    final storeName = isAppleDevice ? 'App Store' : 'Play Store';

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (bottomSheetCtx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: const Color(0xFFE2E8F0),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 16),
              const Text('⭐', style: TextStyle(fontSize: 40)),
              const SizedBox(height: 12),
              const Text(
                'Enjoying Zippyra?',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w900, color: Color(0xFF0F172A)),
              ),
              const SizedBox(height: 6),
              Text(
                'Rate us on the $storeName — it takes 10 seconds!',
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 12, color: Color(0xFF64748B)),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(
                  5,
                  (index) => const Padding(
                    padding: EdgeInsets.symmetric(horizontal: 4),
                    child: Text('⭐', style: TextStyle(fontSize: 26)),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFD97706),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  onPressed: () {
                    Navigator.pop(bottomSheetCtx);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text('Thank you for rating Zippyra on $storeName! ⭐'),
                        backgroundColor: const Color(0xFF16A34A),
                      ),
                    );
                  },
                  child: Text('Rate on $storeName ⭐', style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 15)),
                ),
              ),
              const SizedBox(height: 10),
              SizedBox(
                width: double.infinity,
                child: TextButton(
                  style: TextButton.styleFrom(
                    backgroundColor: const Color(0xFFEFF6FF),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  onPressed: () => Navigator.pop(bottomSheetCtx),
                  child: const Text('Not Now', style: TextStyle(color: Color(0xFF1E3A8A), fontWeight: FontWeight.w800, fontSize: 15)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: SingleChildScrollView(
          child: Column(
            children: [
              // 1. Blue Top Hero Banner (Figma Screen A-1)
              Container(
                width: double.infinity,
                padding: const EdgeInsets.fromLTRB(20, 20, 20, 24),
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    colors: [Color(0xFF1E3A8A), Color(0xFF0F172A)],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.vertical(bottom: Radius.circular(24)),
                ),
                child: Row(
                  children: [
                    Container(
                      width: 60,
                      height: 60,
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white38, width: 2),
                      ),
                      child: const Icon(Icons.person_rounded, color: Colors.white, size: 36),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _displayName,
                            style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 20, color: Colors.white),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            _phone,
                            style: const TextStyle(fontSize: 13, color: Colors.white70, fontWeight: FontWeight.w500),
                          ),
                          const SizedBox(height: 8),
                          GestureDetector(
                            onTap: () async {
                              final updated = await context.push('/profile/edit', extra: {
                                'name': _displayName,
                                'phone': _phone,
                                'email': _email,
                                'birthday': _birthday,
                                'language': _language,
                              });
                              if (updated is Map) {
                                setState(() {
                                  _displayName = updated['name'] ?? _displayName;
                                  _phone = updated['phone'] ?? _phone;
                                  _email = updated['email'] ?? _email;
                                  _birthday = updated['birthday'] ?? _birthday;
                                  _language = updated['language'] ?? _language;
                                });
                              } else {
                                _loadProfileData();
                              }
                            },
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
                              decoration: BoxDecoration(
                                color: Colors.white.withOpacity(0.2),
                                borderRadius: BorderRadius.circular(20),
                                border: Border.all(color: Colors.white54),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: const [
                                  Text(
                                    'Edit Profile',
                                    style: TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold),
                                  ),
                                  SizedBox(width: 4),
                                  Text('✏️', style: TextStyle(fontSize: 10)),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // 2. Stats 3-Card Row (Figma Screen A-1)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  children: [
                    Expanded(
                      child: _buildStatCard(
                        title: '0',
                        subtitle: 'Orders',
                        titleColor: const Color(0xFF1E3A8A),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _buildStatCard(
                        title: '392',
                        subtitle: 'Points',
                        titleColor: const Color(0xFFD97706),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _buildStatCard(
                        title: 'Gold 🥇',
                        subtitle: 'Tier',
                        titleColor: const Color(0xFF16A34A),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 20),

              // 3. Menu Options List (Figma Screen A-1)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Container(
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: const Color(0xFFE2E8F0)),
                    boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
                  ),
                  child: Column(
                    children: [
                      _buildMenuItem(
                        icon: '📦',
                        title: 'My Orders',
                        subtitle: 'History & Receipts',
                        onTap: () => context.push('/orders'),
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '⭐',
                        title: 'Zippy Points',
                        subtitle: '392 pts · Gold',
                        onTap: () => context.push('/loyalty'),
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '💬',
                        title: 'WhatsApp Receipts',
                        subtitle: 'All invoices',
                        onTap: () {},
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '📍',
                        title: 'Saved Addresses',
                        subtitle: '2 locations',
                        onTap: () {},
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '⚙️',
                        title: 'Settings',
                        subtitle: 'App preferences & notifications',
                        onTap: () => context.push('/settings'),
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '🛡️',
                        title: 'Manage My Data',
                        subtitle: 'Privacy & GDPR/DPDPA controls',
                        onTap: () => context.push('/profile/data'),
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '🌟',
                        title: 'Rate Zippyra App',
                        subtitle: 'Give us feedback on store',
                        onTap: _showAppRatingModal,
                      ),
                      const Divider(height: 1, indent: 64, endIndent: 16, color: Color(0xFFF1F5F9)),
                      _buildMenuItem(
                        icon: '🚪',
                        title: 'Logout',
                        subtitle: 'Sign out of account',
                        textColor: const Color(0xFFDC2626),
                        onTap: _showLogoutConfirmation,
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 24),
              const Text('Zippyra Customer App v1.0.0', style: TextStyle(fontSize: 11, color: Color(0xFF94A3B8))),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
      bottomNavigationBar: Container(
        height: 64,
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: Color(0xFFE5E7EB))),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildBottomNavItem(icon: Icons.home_rounded, label: 'Home', isActive: false, onTap: () => context.go('/home')),
            _buildBottomNavItem(icon: Icons.grid_view_rounded, label: 'Categ.', isActive: false, onTap: () => context.push('/categories')),
            GestureDetector(
              onTap: () => context.push('/scan'),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Transform.translate(
                    offset: const Offset(0, -6),
                    child: Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(16),
                        gradient: const LinearGradient(
                          colors: [Color(0xFF1E3A8A), Color(0xFF0F172A)],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: const Color(0xFF1E3A8A).withOpacity(0.3),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: const Icon(Icons.qr_code_scanner_rounded, color: Colors.white, size: 24),
                    ),
                  ),
                ],
              ),
            ),
            _buildBottomNavItem(icon: Icons.receipt_long_rounded, label: 'Orders', isActive: false, onTap: () => context.push('/orders')),
            _buildBottomNavItem(icon: Icons.person_rounded, label: 'Profile', isActive: true, onTap: () {}),
          ],
        ),
      ),
    );
  }

  Widget _buildStatCard({required String title, required String subtitle, required Color titleColor}) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
      ),
      child: Column(
        children: [
          Text(title, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w900, color: titleColor)),
          const SizedBox(height: 2),
          Text(subtitle, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: Color(0xFF64748B))),
        ],
      ),
    );
  }

  Widget _buildMenuItem({
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

  Widget _buildBottomNavItem({
    required IconData icon,
    required String label,
    required bool isActive,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: isActive ? const Color(0xFF1E3A8A) : const Color(0xFF9CA3AF), size: 22),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: isActive ? FontWeight.bold : FontWeight.w500,
              color: isActive ? const Color(0xFF1E3A8A) : const Color(0xFF9CA3AF),
            ),
          ),
        ],
      ),
    );
  }
}

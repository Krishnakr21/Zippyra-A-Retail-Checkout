import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

import 'package:zippyra_core/zippyra_core.dart';
import '../../../injection_container.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  @override
  void initState() {
    super.initState();
    _checkAppStartRoute();
  }

  Future<void> _checkAppStartRoute() async {
    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    final storage = sl<SecureStorage>();
    final hasSeenOnboarding = await storage.read(key: 'has_seen_onboarding');

    if (hasSeenOnboarding != 'true') {
      context.go('/onboarding');
      return;
    }

    final accessToken = await storage.read(key: 'access_token');
    if (accessToken != null && accessToken.isNotEmpty) {
      context.go('/home');
    } else {
      context.go('/auth');
    }
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      backgroundColor: ZippyraColors.primaryBlue,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.shopping_bag_outlined, size: 80, color: Colors.white),
            SizedBox(height: 16),
            Text(
              'Zippyra',
              style: TextStyle(
                color: Colors.white,
                fontSize: 32,
                fontWeight: FontWeight.w800,
                letterSpacing: 1.2,
              ),
            ),
            SizedBox(height: 8),
            Text(
              'Self-Checkout & Anti-Theft Platform',
              style: TextStyle(color: Colors.white70, fontSize: 14),
            ),
            SizedBox(height: 32),
            CircularProgressIndicator(color: Colors.white),
          ],
        ),
      ),
    );
  }
}

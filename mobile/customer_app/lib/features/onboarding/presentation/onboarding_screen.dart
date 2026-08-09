import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import '../../../injection_container.dart';

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({super.key});

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final PageController _pageController = PageController();
  int _currentPage = 0;

  final List<Map<String, String>> _slides = [
    {
      'title': 'Scan Barcode',
      'subtitle': 'Walk into any Zippyra store, scan product barcodes instantly as you pick items.',
      'icon': 'qr_code_scanner',
    },
    {
      'title': 'Instant Payment',
      'subtitle': 'Pay seamlessly with Razorpay UPI, Cards, or Wallet directly from your phone.',
      'icon': 'payment',
    },
    {
      'title': 'Gate Auto-Unlock',
      'subtitle': 'Show single-use Ed25519 exit QR at turnstile gate and walk out cleanly.',
      'icon': 'door_sliding_outlined',
    },
  ];

  Future<void> _completeOnboarding() async {
    await sl<SecureStorage>().write(key: 'has_seen_onboarding', value: 'true');
    if (mounted) {
      context.go('/auth');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Align(
              alignment: Alignment.topRight,
              child: TextButton(
                onPressed: _completeOnboarding,
                child: const Text('Skip', style: TextStyle(color: ZippyraColors.textSecondary)),
              ),
            ),
            Expanded(
              child: PageView.builder(
                controller: _pageController,
                onPageChanged: (index) => setState(() => _currentPage = index),
                itemCount: _slides.length,
                itemBuilder: (context, index) {
                  return Padding(
                    padding: const EdgeInsets.all(32.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Container(
                          padding: const EdgeInsets.all(28),
                          decoration: BoxDecoration(
                            color: ZippyraColors.primaryBlue.withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            index == 0
                                ? Icons.qr_code_scanner
                                : index == 1
                                    ? Icons.payment
                                    : Icons.door_sliding_outlined,
                            size: 80,
                            color: ZippyraColors.primaryBlue,
                          ),
                        ),
                        const SizedBox(height: 32),
                        Text(
                          _slides[index]['title']!,
                          style: const TextStyle(
                            fontSize: 26,
                            fontWeight: FontWeight.w800,
                            color: ZippyraColors.textPrimary,
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          _slides[index]['subtitle']!,
                          textAlign: TextAlign.center,
                          style: const TextStyle(
                            fontSize: 15,
                            color: ZippyraColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(
                _slides.length,
                (i) => Container(
                  margin: const EdgeInsets.symmetric(horizontal: 4),
                  width: _currentPage == i ? 24 : 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: _currentPage == i ? ZippyraColors.primaryBlue : ZippyraColors.border,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 8.0),
              child: ZButton(
                label: _currentPage == _slides.length - 1 ? 'Get Started' : 'Next',
                onPressed: () {
                  if (_currentPage < _slides.length - 1) {
                    _pageController.nextPage(
                      duration: const Duration(milliseconds: 300),
                      curve: Curves.easeInOut,
                    );
                  } else {
                    _completeOnboarding();
                  }
                },
              ),
            ),
            Padding(
              padding: const EdgeInsets.only(bottom: 16.0),
              child: GestureDetector(
                onTap: () {
                  showDialog(
                    context: context,
                    builder: (ctx) => AlertDialog(
                      title: const Text('Legal & Compliance'),
                      content: const Text('By continuing, you agree to Zippyra Terms of Service (v1.2) and Privacy Policy (DPDP Act 2023 compliant). View full documents at zippyra.com/legal'),
                      actions: [
                        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close')),
                      ],
                    ),
                  );
                },
                child: const Text.rich(
                  TextSpan(
                    text: 'By continuing you agree to our ',
                    style: TextStyle(fontSize: 11, color: Colors.grey),
                    children: [
                      TextSpan(
                        text: 'Terms of Service',
                        style: TextStyle(fontWeight: FontWeight.bold, decoration: TextDecoration.underline),
                      ),
                      TextSpan(text: ' & '),
                      TextSpan(
                        text: 'Privacy Policy',
                        style: TextStyle(fontWeight: FontWeight.bold, decoration: TextDecoration.underline),
                      ),
                    ],
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

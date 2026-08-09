import 'dart:ui';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../core/utils/google_auth_launcher.dart';
import '../../../injection_container.dart';
import '../../loyalty/domain/repositories/referral_repository.dart';
import 'bloc/auth_bloc.dart';
import 'widgets/google_signin_button.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> with SingleTickerProviderStateMixin {
  int _selectedSegment = 0; // 0 for Phone, 1 for Email
  bool _showReferralInput = false;
  final TextEditingController _phoneController = TextEditingController();
  final TextEditingController _emailController = TextEditingController();
  final TextEditingController _referralController = TextEditingController();

  late AnimationController _animController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 700),
    );
    _fadeAnimation = CurvedAnimation(parent: _animController, curve: Curves.easeOut);
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 0.1),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _animController, curve: Curves.easeOutCubic));

    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    _referralController.dispose();
    super.dispose();
  }

  void _onContinuePressed() {
    final String channel = _selectedSegment == 0 ? 'phone' : 'email';
    final String identifier = _selectedSegment == 0
        ? _phoneController.text.trim()
        : _emailController.text.trim();

    if (identifier.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Please enter your ${channel == 'phone' ? 'phone number' : 'email address'}'),
          behavior: SnackBarBehavior.floating,
          backgroundColor: const Color(0xFF1E293B),
        ),
      );
      return;
    }

    context.read<AuthBloc>().add(
          SendOtpRequested(channel: channel, identifier: identifier),
        );
  }

  void _onGoogleSignInPressed() {
    context.read<AuthBloc>().add(GoogleSignInRequested());
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is OtpSent) {
          context.push('/auth/otp', extra: {
            'channel': state.channel,
            'identifier': state.identifier,
          });
        } else if (state is AuthSuccess) {
          final refCode = _referralController.text.trim();
          if (refCode.isNotEmpty) {
            sl<ReferralRepository>().applyReferralCode(refCode).catchError((_) {});
          }
          context.go('/permissions');
        } else if (state is AuthFailureState) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(state.message),
              backgroundColor: ZippyraColors.errorRed,
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      },
      child: Scaffold(
        backgroundColor: const Color(0xFF0F172A), // Modern rich dark background
        body: Stack(
          children: [
            // Ambient glowing background orbs
            Positioned(
              top: -80,
              right: -60,
              child: Container(
                width: 260,
                height: 260,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF3B82F6).withOpacity(0.35),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),
            Positioned(
              bottom: -100,
              left: -80,
              child: Container(
                width: 320,
                height: 320,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF6366F1).withOpacity(0.3),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),

            SafeArea(
              child: Center(
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
                  child: SlideTransition(
                    position: _slideAnimation,
                    child: FadeTransition(
                      opacity: _fadeAnimation,
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          // App Logo & Header Badge
                          Container(
                            padding: const EdgeInsets.all(14),
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              gradient: const LinearGradient(
                                colors: [Color(0xFF3B82F6), Color(0xFF1D4ED8)],
                                begin: Alignment.topLeft,
                                end: Alignment.bottomRight,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(0xFF3B82F6).withOpacity(0.4),
                                  blurRadius: 20,
                                  offset: const Offset(0, 8),
                                ),
                              ],
                            ),
                            child: const Icon(
                              Icons.bolt_rounded,
                              size: 40,
                              color: Colors.white,
                            ),
                          ),
                          const SizedBox(height: 16),
                          const Text(
                            'ZIPPYRA',
                            style: TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.w900,
                              letterSpacing: 2,
                              color: Colors.white,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Text(
                            'Instant Retail & Self-Checkout',
                            style: TextStyle(
                              fontSize: 14,
                              color: Colors.white.withOpacity(0.7),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 32),

                          // Glassmorphic Auth Card
                          ClipRRect(
                            borderRadius: BorderRadius.circular(24),
                            child: BackdropFilter(
                              filter: ImageFilter.blur(sigmaX: 16, sigmaY: 16),
                              child: Container(
                                padding: const EdgeInsets.all(24),
                                decoration: BoxDecoration(
                                  color: Colors.white.withOpacity(0.08),
                                  borderRadius: BorderRadius.circular(24),
                                  border: Border.all(
                                    color: Colors.white.withOpacity(0.15),
                                    width: 1.5,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: Colors.black.withOpacity(0.3),
                                      blurRadius: 30,
                                      offset: const Offset(0, 10),
                                    ),
                                  ],
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.stretch,
                                  children: [
                                    // Animated Segment Switcher
                                    Container(
                                      padding: const EdgeInsets.all(4),
                                      decoration: BoxDecoration(
                                        color: Colors.black.withOpacity(0.25),
                                        borderRadius: BorderRadius.circular(14),
                                        border: Border.all(
                                          color: Colors.white.withOpacity(0.08),
                                        ),
                                      ),
                                      child: Row(
                                        children: [
                                          Expanded(
                                            child: GestureDetector(
                                              onTap: () => setState(() => _selectedSegment = 0),
                                              child: AnimatedContainer(
                                                duration: const Duration(milliseconds: 250),
                                                curve: Curves.easeInOut,
                                                padding: const EdgeInsets.symmetric(vertical: 12),
                                                decoration: BoxDecoration(
                                                  gradient: _selectedSegment == 0
                                                      ? const LinearGradient(
                                                          colors: [Color(0xFF3B82F6), Color(0xFF2563EB)],
                                                        )
                                                      : null,
                                                  color: _selectedSegment == 0 ? null : Colors.transparent,
                                                  borderRadius: BorderRadius.circular(10),
                                                  boxShadow: _selectedSegment == 0
                                                      ? [
                                                          BoxShadow(
                                                            color: const Color(0xFF3B82F6).withOpacity(0.3),
                                                            blurRadius: 10,
                                                          )
                                                        ]
                                                      : [],
                                                ),
                                                child: Row(
                                                  mainAxisAlignment: MainAxisAlignment.center,
                                                  children: [
                                                    Icon(
                                                      Icons.phone_iphone_rounded,
                                                      size: 18,
                                                      color: _selectedSegment == 0
                                                          ? Colors.white
                                                          : Colors.white.withOpacity(0.5),
                                                    ),
                                                    const SizedBox(width: 8),
                                                    Text(
                                                      'Phone',
                                                      style: TextStyle(
                                                        fontWeight: FontWeight.bold,
                                                        fontSize: 14,
                                                        color: _selectedSegment == 0
                                                            ? Colors.white
                                                            : Colors.white.withOpacity(0.6),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                          ),
                                          Expanded(
                                            child: GestureDetector(
                                              onTap: () => setState(() => _selectedSegment = 1),
                                              child: AnimatedContainer(
                                                duration: const Duration(milliseconds: 250),
                                                curve: Curves.easeInOut,
                                                padding: const EdgeInsets.symmetric(vertical: 12),
                                                decoration: BoxDecoration(
                                                  gradient: _selectedSegment == 1
                                                      ? const LinearGradient(
                                                          colors: [Color(0xFF3B82F6), Color(0xFF2563EB)],
                                                        )
                                                      : null,
                                                  color: _selectedSegment == 1 ? null : Colors.transparent,
                                                  borderRadius: BorderRadius.circular(10),
                                                  boxShadow: _selectedSegment == 1
                                                      ? [
                                                          BoxShadow(
                                                            color: const Color(0xFF3B82F6).withOpacity(0.3),
                                                            blurRadius: 10,
                                                          )
                                                        ]
                                                      : [],
                                                ),
                                                child: Row(
                                                  mainAxisAlignment: MainAxisAlignment.center,
                                                  children: [
                                                    Icon(
                                                      Icons.alternate_email_rounded,
                                                      size: 18,
                                                      color: _selectedSegment == 1
                                                          ? Colors.white
                                                          : Colors.white.withOpacity(0.5),
                                                    ),
                                                    const SizedBox(width: 8),
                                                    Text(
                                                      'Email',
                                                      style: TextStyle(
                                                        fontWeight: FontWeight.bold,
                                                        fontSize: 14,
                                                        color: _selectedSegment == 1
                                                            ? Colors.white
                                                            : Colors.white.withOpacity(0.6),
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    const SizedBox(height: 24),

                                    // Header Title
                                    AnimatedSwitcher(
                                      duration: const Duration(milliseconds: 200),
                                      child: Column(
                                        key: ValueKey<int>(_selectedSegment),
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            _selectedSegment == 0
                                                ? 'Enter Mobile Number'
                                                : 'Enter Email Address',
                                            style: const TextStyle(
                                              fontSize: 20,
                                              fontWeight: FontWeight.bold,
                                              color: Colors.white,
                                            ),
                                          ),
                                          const SizedBox(height: 4),
                                          Text(
                                            _selectedSegment == 0
                                                ? 'We will send a 6-digit verification code to your phone'
                                                : 'We will send a 6-digit verification code to your email',
                                            style: TextStyle(
                                              fontSize: 13,
                                              color: Colors.white.withOpacity(0.6),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    const SizedBox(height: 20),

                                    // Input Field
                                    if (_selectedSegment == 0)
                                      _buildGlassInput(
                                        controller: _phoneController,
                                        hint: '+91 98765 43210',
                                        keyboardType: TextInputType.phone,
                                        icon: Icons.phone_android_rounded,
                                      )
                                    else
                                      _buildGlassInput(
                                        controller: _emailController,
                                        hint: 'you@example.com',
                                        keyboardType: TextInputType.emailAddress,
                                        icon: Icons.email_outlined,
                                      ),

                                    const SizedBox(height: 12),

                                    // Referral toggle
                                    GestureDetector(
                                      onTap: () => setState(() => _showReferralInput = !_showReferralInput),
                                      child: Padding(
                                        padding: const EdgeInsets.symmetric(vertical: 4),
                                        child: Row(
                                          children: [
                                            Icon(
                                              _showReferralInput
                                                  ? Icons.remove_circle_outline_rounded
                                                  : Icons.card_giftcard_rounded,
                                              size: 16,
                                              color: const Color(0xFF60A5FA),
                                            ),
                                            const SizedBox(width: 6),
                                            Text(
                                              _showReferralInput ? 'Hide Referral Code' : 'Have a referral code?',
                                              style: const TextStyle(
                                                color: Color(0xFF60A5FA),
                                                fontWeight: FontWeight.w600,
                                                fontSize: 13,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ),

                                    if (_showReferralInput) ...[
                                      const SizedBox(height: 10),
                                      _buildGlassInput(
                                        controller: _referralController,
                                        hint: 'Referral Code (e.g. REF98A2K)',
                                        keyboardType: TextInputType.text,
                                        icon: Icons.confirmation_number_outlined,
                                      ),
                                    ],

                                    const SizedBox(height: 24),

                                    // Action Button
                                    BlocBuilder<AuthBloc, AuthState>(
                                      builder: (context, state) {
                                        final isLoading = state is AuthLoading;
                                        return _buildGradientButton(
                                          label: 'Get OTP Code',
                                          isLoading: isLoading,
                                          onPressed: _onContinuePressed,
                                        );
                                      },
                                    ),

                                    const SizedBox(height: 20),

                                    // Divider "OR"
                                    Row(
                                      children: [
                                        Expanded(child: Container(height: 1, color: Colors.white.withOpacity(0.12))),
                                        Padding(
                                          padding: const EdgeInsets.symmetric(horizontal: 14),
                                          child: Text(
                                            'OR',
                                            style: TextStyle(
                                              color: Colors.white.withOpacity(0.4),
                                              fontSize: 12,
                                              fontWeight: FontWeight.bold,
                                              letterSpacing: 1,
                                            ),
                                          ),
                                        ),
                                        Expanded(child: Container(height: 1, color: Colors.white.withOpacity(0.12))),
                                      ],
                                    ),

                                    const SizedBox(height: 20),

                                    // Google Sign-In Button
                                    BlocBuilder<AuthBloc, AuthState>(
                                      builder: (context, state) {
                                        final isGoogleLoading = state is AuthGoogleInProgress;
                                        return GoogleSignInButton(
                                          isLoading: isGoogleLoading,
                                          onPressed: _onGoogleSignInPressed,
                                        );
                                      },
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildGlassInput({
    required TextEditingController controller,
    required String hint,
    required TextInputType keyboardType,
    required IconData icon,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.06),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: Colors.white.withOpacity(0.15),
        ),
      ),
      child: TextField(
        controller: controller,
        keyboardType: keyboardType,
        style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: TextStyle(color: Colors.white.withOpacity(0.35)),
          prefixIcon: Icon(icon, color: const Color(0xFF60A5FA)),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        ),
      ),
    );
  }

  Widget _buildGradientButton({
    required String label,
    required bool isLoading,
    required VoidCallback onPressed,
  }) {
    return Container(
      height: 52,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(14),
        gradient: const LinearGradient(
          colors: [Color(0xFF3B82F6), Color(0xFF1D4ED8)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF3B82F6).withOpacity(0.4),
            blurRadius: 16,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: ElevatedButton(
        onPressed: isLoading ? null : onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.transparent,
          shadowColor: Colors.transparent,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
        ),
        child: isLoading
            ? const SizedBox(
                width: 22,
                height: 22,
                child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5),
              )
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    label,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Icon(Icons.arrow_forward_rounded, color: Colors.white, size: 20),
                ],
              ),
      ),
    );
  }
}


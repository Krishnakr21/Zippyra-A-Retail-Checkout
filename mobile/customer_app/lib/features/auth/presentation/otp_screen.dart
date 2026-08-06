import 'dart:async';
import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'bloc/auth_bloc.dart';

class OtpScreen extends StatefulWidget {
  final String channel;
  final String identifier;

  const OtpScreen({
    super.key,
    required this.channel,
    required this.identifier,
  });

  @override
  State<OtpScreen> createState() => _OtpScreenState();
}

class _OtpScreenState extends State<OtpScreen> with SingleTickerProviderStateMixin {
  final TextEditingController _otpController = TextEditingController();
  final FocusNode _focusNode = FocusNode();
  int _resendSeconds = 30;
  Timer? _timer;

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _startResendTimer();

    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 2),
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.95, end: 1.05).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _timer?.cancel();
    _pulseController.dispose();
    _otpController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _startResendTimer() {
    setState(() => _resendSeconds = 30);
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_resendSeconds > 0) {
        setState(() => _resendSeconds--);
      } else {
        timer.cancel();
      }
    });
  }

  String _maskIdentifier(String channel, String id) {
    if (channel == 'phone') {
      if (id.length < 7) return id;
      return '${id.substring(0, 3)}XXXXXX${id.substring(id.length - 4)}';
    } else {
      final parts = id.split('@');
      if (parts.length != 2 || parts[0].isEmpty) return id;
      return '${parts[0][0]}***@${parts[1]}';
    }
  }

  void _onVerifyPressed() {
    final otp = _otpController.text.trim();
    if (otp.length != 6) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter full 6-digit OTP code'),
          behavior: SnackBarBehavior.floating,
          backgroundColor: Color(0xFF1E293B),
        ),
      );
      return;
    }
    context.read<AuthBloc>().add(
          VerifyOtpRequested(channel: widget.channel, identifier: widget.identifier, otp: otp),
        );
  }

  void _onResendPressed() {
    if (_resendSeconds > 0) return;
    context.read<AuthBloc>().add(
          SendOtpRequested(channel: widget.channel, identifier: widget.identifier),
        );
    _startResendTimer();
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('New verification code sent!'),
        behavior: SnackBarBehavior.floating,
        backgroundColor: ZippyraColors.successGreen,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthSuccess) {
          if (state.session.isNewUser) {
            context.go('/profile-completion');
          } else {
            context.go('/permissions');
          }
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
        backgroundColor: const Color(0xFF0F172A),
        body: Stack(
          children: [
            // Background Orbs
            Positioned(
              top: -60,
              left: -60,
              child: Container(
                width: 280,
                height: 280,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF3B82F6).withOpacity(0.3),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),
            Positioned(
              bottom: -80,
              right: -60,
              child: Container(
                width: 300,
                height: 300,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF8B5CF6).withOpacity(0.3),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),

            SafeArea(
              child: Column(
                children: [
                  // Custom Header with Back button
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    child: Row(
                      children: [
                        IconButton(
                          onPressed: () => context.pop(),
                          icon: Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.08),
                              shape: BoxShape.circle,
                              border: Border.all(color: Colors.white.withOpacity(0.12)),
                            ),
                            child: const Icon(Icons.arrow_back_rounded, color: Colors.white, size: 20),
                          ),
                        ),
                        const Spacer(),
                        Text(
                          'SECURITY VERIFICATION',
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                            letterSpacing: 1.5,
                            color: Colors.white.withOpacity(0.6),
                          ),
                        ),
                        const Spacer(),
                        const SizedBox(width: 48),
                      ],
                    ),
                  ),

                  Expanded(
                    child: Center(
                      child: SingleChildScrollView(
                        physics: const BouncingScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            // Animated Shield Icon
                            ScaleTransition(
                              scale: _pulseAnimation,
                              child: Container(
                                padding: const EdgeInsets.all(20),
                                decoration: BoxDecoration(
                                  shape: BoxShape.circle,
                                  gradient: const LinearGradient(
                                    colors: [Color(0xFF3B82F6), Color(0xFF6366F1)],
                                    begin: Alignment.topLeft,
                                    end: Alignment.bottomRight,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: const Color(0xFF3B82F6).withOpacity(0.45),
                                      blurRadius: 24,
                                      offset: const Offset(0, 8),
                                    ),
                                  ],
                                ),
                                child: const Icon(
                                  Icons.shield_outlined,
                                  size: 44,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                            const SizedBox(height: 24),

                            const Text(
                              'Verify Your Identity',
                              style: TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                            const SizedBox(height: 8),
                            RichText(
                              textAlign: TextAlign.center,
                              text: TextSpan(
                                style: TextStyle(color: Colors.white.withOpacity(0.6), fontSize: 14),
                                children: [
                                  const TextSpan(text: 'Enter 6-digit code sent to\n'),
                                  TextSpan(
                                    text: _maskIdentifier(widget.channel, widget.identifier),
                                    style: const TextStyle(
                                      color: Color(0xFF60A5FA),
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 32),

                            // Glassmorphic OTP Card
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
                                    children: [
                                      // Invisible TextField overlaid on styled digit boxes
                                      GestureDetector(
                                        onTap: () => FocusScope.of(context).requestFocus(_focusNode),
                                        child: Stack(
                                          alignment: Alignment.center,
                                          children: [
                                            // Hidden real TextField
                                            Opacity(
                                              opacity: 0.0,
                                              child: TextField(
                                                controller: _otpController,
                                                focusNode: _focusNode,
                                                keyboardType: TextInputType.number,
                                                maxLength: 6,
                                                onChanged: (val) {
                                                  setState(() {});
                                                  if (val.length == 6) {
                                                    _onVerifyPressed();
                                                  }
                                                },
                                              ),
                                            ),

                                            // 6 Styled Digit Boxes
                                            Row(
                                              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                                              children: List.generate(6, (index) {
                                                final code = _otpController.text;
                                                final isFilled = index < code.length;
                                                final isCurrent = index == code.length || (index == 5 && code.length == 6);
                                                final digit = isFilled ? code[index] : '';

                                                return AnimatedContainer(
                                                  duration: const Duration(milliseconds: 200),
                                                  width: 44,
                                                  height: 54,
                                                  alignment: Alignment.center,
                                                  decoration: BoxDecoration(
                                                    color: isFilled
                                                        ? const Color(0xFF3B82F6).withOpacity(0.2)
                                                        : Colors.white.withOpacity(0.06),
                                                    borderRadius: BorderRadius.circular(12),
                                                    border: Border.all(
                                                      color: isCurrent
                                                          ? const Color(0xFF60A5FA)
                                                          : (isFilled
                                                              ? const Color(0xFF3B82F6)
                                                              : Colors.white.withOpacity(0.15)),
                                                      width: isCurrent ? 2 : 1,
                                                    ),
                                                    boxShadow: isCurrent
                                                        ? [
                                                            BoxShadow(
                                                              color: const Color(0xFF3B82F6).withOpacity(0.4),
                                                              blurRadius: 10,
                                                            )
                                                          ]
                                                        : [],
                                                  ),
                                                  child: Text(
                                                    digit,
                                                    style: const TextStyle(
                                                      fontSize: 22,
                                                      fontWeight: FontWeight.bold,
                                                      color: Colors.white,
                                                    ),
                                                  ),
                                                );
                                              }),
                                            ),
                                          ],
                                        ),
                                      ),
                                      const SizedBox(height: 24),

                                      // Resend Counter & Button
                                      Row(
                                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                        children: [
                                          Row(
                                            children: [
                                              Icon(
                                                Icons.timer_outlined,
                                                size: 16,
                                                color: Colors.white.withOpacity(0.5),
                                              ),
                                              const SizedBox(width: 6),
                                              Text(
                                                _resendSeconds > 0 ? 'Resend in ${_resendSeconds}s' : 'Didn\'t receive code?',
                                                style: TextStyle(
                                                  color: Colors.white.withOpacity(0.6),
                                                  fontSize: 13,
                                                ),
                                              ),
                                            ],
                                          ),
                                          TextButton(
                                            onPressed: _resendSeconds == 0 ? _onResendPressed : null,
                                            style: TextButton.styleFrom(
                                              padding: EdgeInsets.zero,
                                              minimumSize: Size.zero,
                                              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                            ),
                                            child: Text(
                                              'Resend OTP',
                                              style: TextStyle(
                                                color: _resendSeconds == 0
                                                    ? const Color(0xFF60A5FA)
                                                    : Colors.white.withOpacity(0.3),
                                                fontWeight: FontWeight.bold,
                                                fontSize: 13,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),

                                      const SizedBox(height: 28),

                                      // Verify Button
                                      BlocBuilder<AuthBloc, AuthState>(
                                        builder: (context, state) {
                                          final isLoading = state is AuthLoading;
                                          return _buildVerifyButton(
                                            isLoading: isLoading,
                                            onPressed: _onVerifyPressed,
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
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerifyButton({
    required bool isLoading,
    required VoidCallback onPressed,
  }) {
    final isReady = _otpController.text.length == 6;

    return Container(
      height: 52,
      width: double.infinity,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(14),
        gradient: LinearGradient(
          colors: isReady
              ? [const Color(0xFF3B82F6), const Color(0xFF1D4ED8)]
              : [const Color(0xFF1E293B), const Color(0xFF334155)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        boxShadow: isReady
            ? [
                BoxShadow(
                  color: const Color(0xFF3B82F6).withOpacity(0.4),
                  blurRadius: 16,
                  offset: const Offset(0, 6),
                ),
              ]
            : [],
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
                children: const [
                  Text(
                    'Verify & Proceed',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 0.5,
                    ),
                  ),
                  SizedBox(width: 8),
                  Icon(Icons.check_circle_outline_rounded, color: Colors.white, size: 20),
                ],
              ),
      ),
    );
  }
}


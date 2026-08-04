import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'presentation/bloc/auth_bloc.dart';
import 'presentation/screens/pin_setup_screen.dart';
import 'presentation/widgets/pin_pad.dart';

class AuthScreen extends StatefulWidget {
  const AuthScreen({super.key});

  @override
  State<AuthScreen> createState() => _AuthScreenState();
}

class _AuthScreenState extends State<AuthScreen> {
  bool _isPinMode = false;
  final TextEditingController _phoneController = TextEditingController(text: '+919876543210');
  final TextEditingController _otpController = TextEditingController();
  bool _otpSent = false;
  String? _infoBannerMessage;
  int _lockoutSeconds = 0;
  Timer? _lockoutTimer;

  @override
  void dispose() {
    _phoneController.dispose();
    _otpController.dispose();
    _lockoutTimer?.cancel();
    super.dispose();
  }

  void _startLockoutTimer(int seconds) {
    _lockoutTimer?.cancel();
    setState(() {
      _lockoutSeconds = seconds;
    });
    _lockoutTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_lockoutSeconds <= 1) {
        timer.cancel();
        setState(() {
          _lockoutSeconds = 0;
        });
      } else {
        setState(() {
          _lockoutSeconds--;
        });
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthOtpSent) {
          setState(() {
            _otpSent = true;
            _infoBannerMessage = null;
          });
        } else if (state is AuthStaffNotRegistered) {
          setState(() {
            _infoBannerMessage = "This number isn't registered as staff. Ask your store manager to add you.";
          });
        } else if (state is AuthPinNotSet) {
          setState(() {
            _isPinMode = false;
            _infoBannerMessage = "No PIN configured yet. Falling back to OTP login.";
          });
        } else if (state is AuthPinLocked) {
          _startLockoutTimer(state.retryAfterSeconds);
        } else if (state is AuthAuthenticated) {
          if (!state.session.hasPinSet) {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => PinSetupScreen(
                  onCompleted: () => Navigator.of(context).pop(),
                ),
              ),
            );
          }
        }
      },
      builder: (context, state) {
        return Scaffold(
          body: SafeArea(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const SizedBox(height: 40),
                  const Icon(Icons.storefront_rounded, size: 64, color: Colors.indigo),
                  const SizedBox(height: 16),
                  const Text(
                    'Zippyra Staff App',
                    style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _isPinMode ? 'Enter your Phone & Quick PIN' : 'Enter registered mobile number',
                    style: TextStyle(color: Colors.grey.shade600),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 24),

                  if (_infoBannerMessage != null) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.orange.shade50,
                        border: Border.all(color: Colors.orange.shade300),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        _infoBannerMessage!,
                        style: TextStyle(color: Colors.orange.shade900, fontSize: 13),
                        textAlign: TextAlign.center,
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],

                  // Phone Input Field
                  TextField(
                    key: const Key('phone_input_field'),
                    controller: _phoneController,
                    keyboardType: TextInputType.phone,
                    decoration: const InputDecoration(
                      labelText: 'Phone Number',
                      prefixIcon: Icon(Icons.phone_android),
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),

                  if (!_isPinMode) ...[
                    // OTP Mode
                    if (_otpSent) ...[
                      TextField(
                        key: const Key('otp_input_field'),
                        controller: _otpController,
                        keyboardType: TextInputType.number,
                        maxLength: 6,
                        decoration: const InputDecoration(
                          labelText: '6-digit OTP',
                          prefixIcon: Icon(Icons.lock_clock_outlined),
                          border: OutlineInputBorder(),
                        ),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        key: const Key('verify_otp_button'),
                        onPressed: state is AuthLoading
                            ? null
                            : () {
                                context.read<AuthBloc>().add(
                                      AuthVerifyOtpRequested(
                                        _phoneController.text.trim(),
                                        _otpController.text.trim(),
                                      ),
                                    );
                              },
                        child: state is AuthLoading
                            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                            : const Text('Verify OTP & Login'),
                      ),
                    ] else ...[
                      ElevatedButton(
                        key: const Key('send_otp_button'),
                        onPressed: state is AuthLoading
                            ? null
                            : () {
                                context.read<AuthBloc>().add(
                                      AuthSendOtpRequested(_phoneController.text.trim()),
                                    );
                              },
                        child: state is AuthLoading
                            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                            : const Text('Send OTP'),
                      ),
                    ],

                    const SizedBox(height: 16),
                    TextButton(
                      key: const Key('toggle_pin_mode_link'),
                      onPressed: () {
                        setState(() {
                          _isPinMode = true;
                          _infoBannerMessage = null;
                        });
                      },
                      child: const Text('Use PIN instead'),
                    ),
                  ] else ...[
                    // PIN Mode
                    PinPadWidget(
                      key: const Key('pin_pad_widget'),
                      maxLength: 4,
                      isLocked: _lockoutSeconds > 0,
                      lockedText: _lockoutSeconds > 0
                          ? 'PIN locked due to failed attempts. Try again in ${_lockoutSeconds}s or use OTP.'
                          : null,
                      onCompleted: (pin) {
                        context.read<AuthBloc>().add(
                              AuthPinLoginRequested(_phoneController.text.trim(), pin),
                            );
                      },
                    ),
                    const SizedBox(height: 16),
                    TextButton(
                      key: const Key('toggle_otp_mode_link'),
                      onPressed: () {
                        setState(() {
                          _isPinMode = false;
                          _infoBannerMessage = null;
                        });
                      },
                      child: const Text('Use OTP instead'),
                    ),
                  ],
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

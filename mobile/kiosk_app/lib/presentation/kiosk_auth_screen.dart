import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class KioskAuthScreen extends StatefulWidget {
  final Function(String phone) onAuthenticated;
  final VoidCallback onCancel;

  const KioskAuthScreen({
    super.key,
    required this.onAuthenticated,
    required this.onCancel,
  });

  @override
  State<KioskAuthScreen> createState() => _KioskAuthScreenState();
}

class _KioskAuthScreenState extends State<KioskAuthScreen> {
  String _phoneNumber = '';
  String _otpCode = '';
  bool _otpSent = false;
  bool _isLoading = false;

  void _onKeyPress(String digit) {
    if (_isLoading) return;
    setState(() {
      if (!_otpSent) {
        if (_phoneNumber.length < 10) {
          _phoneNumber += digit;
        }
      } else {
        if (_otpCode.length < 6) {
          _otpCode += digit;
        }
      }
    });
  }

  void _onBackspace() {
    if (_isLoading) return;
    setState(() {
      if (!_otpSent) {
        if (_phoneNumber.isNotEmpty) {
          _phoneNumber = _phoneNumber.substring(0, _phoneNumber.length - 1);
        }
      } else {
        if (_otpCode.isNotEmpty) {
          _otpCode = _otpCode.substring(0, _otpCode.length - 1);
        }
      }
    });
  }

  void _onClear() {
    if (_isLoading) return;
    setState(() {
      if (!_otpSent) {
        _phoneNumber = '';
      } else {
        _otpCode = '';
      }
    });
  }

  void _submitPhone() {
    if (_phoneNumber.length != 10) return;
    setState(() {
      _isLoading = true;
    });

    // Simulate OTP send call to auth-service
    Future.delayed(const Duration(milliseconds: 600), () {
      setState(() {
        _isLoading = false;
        _otpSent = true;
      });
    });
  }

  void _verifyOtp() {
    if (_otpCode.length != 6) return;
    setState(() {
      _isLoading = true;
    });

    // Simulate OTP verification call to auth-service
    Future.delayed(const Duration(milliseconds: 600), () {
      widget.onAuthenticated(_phoneNumber);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: ZippyraColors.darkBackground,
      body: SafeArea(
        child: Column(
          children: [
            // Top Nav
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
              color: ZippyraColors.darkCard,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    _otpSent ? 'Step 2: Enter Verification OTP' : 'Step 1: Enter Mobile Number',
                    style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.white),
                  ),
                  TextButton.icon(
                    style: TextButton.styleFrom(foregroundColor: Colors.redAccent),
                    onPressed: widget.onCancel,
                    icon: const Icon(Icons.cancel_outlined, size: 28),
                    label: const Text('Cancel Session', style: TextStyle(fontSize: 18)),
                  ),
                ],
              ),
            ),

            Expanded(
              child: Row(
                children: [
                  // Left Display Panel
                  Expanded(
                    flex: 5,
                    child: Padding(
                      padding: const EdgeInsets.all(48.0),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            _otpSent ? Icons.mark_email_read : Icons.phone_android,
                            size: 80,
                            color: ZippyraColors.primary,
                          ),
                          const SizedBox(height: 24),
                          Text(
                            _otpSent ? 'Enter 6-digit OTP' : 'Enter Your Phone Number',
                            style: const TextStyle(fontSize: 32, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                          const SizedBox(height: 12),
                          Text(
                            _otpSent
                                ? 'SMS OTP sent to +91 $_phoneNumber. Check your phone.'
                                : 'Enter your 10-digit mobile number to access your cart, earn loyalty points, and receive your digital exit pass via SMS.',
                            style: const TextStyle(fontSize: 18, color: Colors.white70),
                          ),
                          const SizedBox(height: 36),

                          // Value Input Display Box
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
                            decoration: BoxDecoration(
                              color: Colors.black54,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(color: ZippyraColors.primary, width: 2),
                            ),
                            child: Text(
                              _otpSent ? _otpCode : (_phoneNumber.isEmpty ? '10-Digit Mobile Number' : _phoneNumber),
                              style: TextStyle(
                                fontSize: 36,
                                fontWeight: FontWeight.bold,
                                letterSpacing: 4,
                                color: (_otpSent ? _otpCode : _phoneNumber).isEmpty ? Colors.white30 : ZippyraColors.primary,
                              ),
                            ),
                          ),
                          const SizedBox(height: 24),

                          if (_isLoading)
                            const CircularProgressIndicator(color: ZippyraColors.primary)
                          else
                            SizedBox(
                              width: double.infinity,
                              height: 60,
                              child: ElevatedButton(
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: ZippyraColors.primary,
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                                ),
                                onPressed: _otpSent ? _verifyOtp : _submitPhone,
                                child: Text(
                                  _otpSent ? 'VERIFY & CONTINUE' : 'SEND OTP',
                                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black),
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  ),

                  // Right Keypad Panel (On-Screen Touch Keypad)
                  Expanded(
                    flex: 4,
                    child: Container(
                      color: ZippyraColors.darkCard,
                      padding: const EdgeInsets.all(32),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          for (var row in [
                            ['1', '2', '3'],
                            ['4', '5', '6'],
                            ['7', '8', '9'],
                            ['CLEAR', '0', '⌫']
                          ])
                            Padding(
                              padding: const EdgeInsets.symmetric(vertical: 8.0),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                                children: row.map((key) {
                                  return SizedBox(
                                    width: 100,
                                    height: 80,
                                    child: ElevatedButton(
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: key == 'CLEAR'
                                            ? Colors.red.withOpacity(0.2)
                                            : key == '⌫'
                                                ? Colors.orange.withOpacity(0.2)
                                                : Colors.white12,
                                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                                      ),
                                      onPressed: () {
                                        if (key == 'CLEAR') {
                                          _onClear();
                                        } else if (key == '⌫') {
                                          _onBackspace();
                                        } else {
                                          _onKeyPress(key);
                                        }
                                      },
                                      child: Text(
                                        key,
                                        style: TextStyle(
                                          fontSize: 28,
                                          fontWeight: FontWeight.bold,
                                          color: key == 'CLEAR'
                                              ? Colors.redAccent
                                              : key == '⌫'
                                                  ? Colors.orangeAccent
                                                  : Colors.white,
                                        ),
                                      ),
                                    ),
                                  );
                                }).toList(),
                              ),
                            ),
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
    );
  }
}

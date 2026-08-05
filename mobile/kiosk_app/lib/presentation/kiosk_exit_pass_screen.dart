import 'dart:async';
import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class KioskExitPassScreen extends StatefulWidget {
  final String orderId;
  final String exitQr;
  final String customerPhone;
  final VoidCallback onDone;

  const KioskExitPassScreen({
    super.key,
    required this.orderId,
    required this.exitQr,
    required this.customerPhone,
    required this.onDone,
  });

  @override
  State<KioskExitPassScreen> createState() => _KioskExitPassScreenState();
}

class _KioskExitPassScreenState extends State<KioskExitPassScreen> {
  int _secondsRemaining = 15;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_secondsRemaining > 1) {
        setState(() {
          _secondsRemaining--;
        });
      } else {
        _timer?.cancel();
        widget.onDone();
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: ZippyraColors.darkBackground,
      body: SafeArea(
        child: Center(
          child: Container(
            width: 700,
            padding: const EdgeInsets.all(40),
            decoration: BoxDecoration(
              color: ZippyraColors.darkCard,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: ZippyraColors.primary, width: 2),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.check_circle_outline, color: ZippyraColors.primary, size: 90),
                const SizedBox(height: 16),
                const Text('Payment Successful!', style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold, color: Colors.white)),
                const SizedBox(height: 8),
                Text('Order ID: ${widget.orderId}', style: const TextStyle(fontSize: 18, color: Colors.white70)),
                const SizedBox(height: 24),

                // SMS Notification Delivery Confirmation Badge
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                  decoration: BoxDecoration(
                    color: ZippyraColors.primary.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(30),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.sms, color: ZippyraColors.primary, size: 24),
                      const SizedBox(width: 10),
                      Text(
                        'Exit Pass & Receipt SMS Sent to +91 ${widget.customerPhone}',
                        style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 28),

                // QR Code
                QrImageView(
                  data: widget.exitQr,
                  version: QrVersions.auto,
                  size: 200.0,
                  backgroundColor: Colors.white,
                  padding: const EdgeInsets.all(12),
                ),
                const SizedBox(height: 24),
                const Text(
                  'Scan this Exit Pass at the RFID Store Exit Gate',
                  style: TextStyle(fontSize: 16, color: Colors.white70),
                ),
                const SizedBox(height: 32),

                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: ZippyraColors.primary,
                    padding: const EdgeInsets.symmetric(horizontal: 40, vertical: 16),
                  ),
                  onPressed: widget.onDone,
                  child: Text(
                    'DONE (Auto-reset in ${_secondsRemaining}s)',
                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.black),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

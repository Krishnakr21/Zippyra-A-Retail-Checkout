import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/exit_bloc.dart';

class ExitQrScreen extends StatefulWidget {
  final String orderId;
  final String token;
  final DateTime expiresAt;

  const ExitQrScreen({
    super.key,
    required this.orderId,
    required this.token,
    required this.expiresAt,
  });

  @override
  State<ExitQrScreen> createState() => _ExitQrScreenState();
}

class _ExitQrScreenState extends State<ExitQrScreen> {
  @override
  void initState() {
    super.initState();
    context.read<ExitBloc>().add(ExitScreenOpened(
          orderId: widget.orderId,
          token: widget.token,
          expiresAt: widget.expiresAt,
        ));
  }

  String _formatTimer(int totalSeconds) {
    final minutes = totalSeconds ~/ 60;
    final seconds = totalSeconds % 60;
    final mm = minutes.toString().padLeft(2, '0');
    final ss = seconds.toString().padLeft(2, '0');
    return '$mm:$ss';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Store Exit Pass'),
        automaticallyImplyLeading: false,
      ),
      body: BlocConsumer<ExitBloc, ExitState>(
        listener: (context, state) {
          if (state is ExitOpened) {
            context.go('/exit/success');
          } else if (state is ExitTokenExpired) {
            context.go('/exit/expired');
          } else if (state is ExitHelpNeeded) {
            context.go('/exit/help-needed');
          }
        },
        builder: (context, state) {
          String token = widget.token;
          int remainingSeconds = widget.expiresAt.difference(DateTime.now()).inSeconds.clamp(0, 600);
          bool isAwaitingRfid = false;

          if (state is ExitDisplayingQr) {
            token = state.token;
            remainingSeconds = state.remainingSeconds;
          } else if (state is ExitAwaitingRfid) {
            token = state.token;
            remainingSeconds = state.remainingSeconds;
            isAwaitingRfid = true;
          }

          final isWarningTimer = remainingSeconds < 120; // Under 2 minutes
          final timerColor = isWarningTimer ? Colors.orange[800] : ZippyraColors.primary;

          return SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text(
                    'Scan at Store Exit Gate',
                    style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Hold your phone up to the gate scanner to unlock the exit door.',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.grey, fontSize: 14),
                  ),
                  const SizedBox(height: 24),

                  // QR Code Card with optional RFID Deactivation Overlay
                  Stack(
                    alignment: Alignment.center,
                    children: [
                      Card(
                        elevation: 4,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                        child: Padding(
                          padding: const EdgeInsets.all(24.0),
                          child: QrImageView(
                            data: token,
                            version: QrVersions.auto,
                            size: 240.0,
                            backgroundColor: Colors.white,
                          ),
                        ),
                      ),
                      if (isAwaitingRfid)
                        Container(
                          width: 240,
                          height: 240,
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.75),
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: const Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              CircularProgressIndicator(color: Colors.white),
                              SizedBox(height: 16),
                              Text(
                                'Deactivating security tags...',
                                textAlign: TextAlign.center,
                                style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),

                  const SizedBox(height: 24),

                  // Countdown Timer Display
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.timer_outlined, color: timerColor, size: 22),
                      const SizedBox(width: 8),
                      Text(
                        'Pass valid for: ${_formatTimer(remainingSeconds)}',
                        key: const Key('exit_timer_text'),
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: timerColor,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

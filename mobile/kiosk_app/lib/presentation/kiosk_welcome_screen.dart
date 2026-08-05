import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class KioskWelcomeScreen extends StatelessWidget {
  final String storeId;
  final String deviceId;
  final VoidCallback onStart;

  const KioskWelcomeScreen({
    super.key,
    required this.storeId,
    required this.deviceId,
    required this.onStart,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: ZippyraColors.darkBackground,
      body: SafeArea(
        child: Column(
          children: [
            // Top Terminal Header
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 24),
              color: ZippyraColors.darkCard,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(Icons.shopping_bag_outlined, color: ZippyraColors.primary, size: 36),
                      const SizedBox(width: 16),
                      Text(
                        'Zippyra Self-Checkout Terminal',
                        style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                    ],
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    decoration: BoxDecoration(
                      color: ZippyraColors.primary.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: ZippyraColors.primary),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.store, color: ZippyraColors.primary, size: 20),
                        const SizedBox(width: 8),
                        Text(
                          'Store: $storeId',
                          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Main Interactive Welcome Banner
            Expanded(
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(32),
                      decoration: BoxDecoration(
                        color: ZippyraColors.primary.withOpacity(0.1),
                        shape: BoxShape.circle,
                      ),
                      child: Icon(Icons.touch_app, size: 100, color: ZippyraColors.primary),
                    ),
                    const SizedBox(height: 32),
                    Text(
                      'Welcome to Zippyra Express',
                      style: Theme.of(context).textTheme.displaySmall?.copyWith(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Fast, Scan-&-Go Checkout Terminal',
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            color: Colors.white70,
                          ),
                    ),
                    const SizedBox(height: 48),

                    // Touch to Start Button
                    SizedBox(
                      width: 400,
                      height: 80,
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: ZippyraColors.primary,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                          elevation: 8,
                        ),
                        onPressed: onStart,
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: const [
                            Text(
                              'TOUCH TO START',
                              style: TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                                color: Colors.black,
                                letterSpacing: 1.5,
                              ),
                            ),
                            SizedBox(width: 16),
                            Icon(Icons.arrow_forward_ios, color: Colors.black, size: 28),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

            // Terminal Footer Info
            Container(
              padding: const EdgeInsets.all(16),
              color: Colors.black45,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Terminal ID: $deviceId',
                    style: const TextStyle(color: Colors.white54, fontSize: 14),
                  ),
                  const Text(
                    'Powered by Zippyra Platform • Protected by RFID Exit Gates',
                    style: TextStyle(color: Colors.white54, fontSize: 14),
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

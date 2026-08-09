import 'package:flutter/material.dart';
import 'package:sentry_flutter/sentry_flutter.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'presentation/kiosk_welcome_screen.dart';
import 'presentation/kiosk_auth_screen.dart';
import 'presentation/kiosk_checkout_screen.dart';
import 'presentation/kiosk_exit_pass_screen.dart';

enum KioskState { welcome, auth, checkout, exitPass }

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await SentryFlutter.init(
    (options) {
      options.dsn = const String.fromEnvironment('SENTRY_DSN_MOBILE', defaultValue: '');
      options.environment = const String.fromEnvironment('ENV', defaultValue: 'production');
    },
    appRunner: () => runApp(const ZippyraKioskApp()),
  );
}

class ZippyraKioskApp extends StatefulWidget {
  const ZippyraKioskApp({super.key});

  @override
  State<ZippyraKioskApp> createState() => _ZippyraKioskAppState();
}

class _ZippyraKioskAppState extends State<ZippyraKioskApp> {
  // Configured Terminal Settings (loaded from local environment / Electron config)
  final String _storeId = const String.fromEnvironment('KIOSK_STORE_ID', defaultValue: 'STORE-BLR-001');
  final String _deviceId = const String.fromEnvironment('KIOSK_DEVICE_ID', defaultValue: 'KIOSK-TERM-001');

  KioskState _currentState = KioskState.welcome;
  String _customerPhone = '';
  String _currentOrderId = '';
  String _currentExitQr = '';

  void _resetToWelcome() {
    setState(() {
      _currentState = KioskState.welcome;
      _customerPhone = '';
      _currentOrderId = '';
      _currentExitQr = '';
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Zippyra Kiosk',
      theme: ZippyraTheme.lightTheme,
      debugShowCheckedModeBanner: false,
      home: _buildCurrentStateScreen(),
    );
  }

  Widget _buildCurrentStateScreen() {
    switch (_currentState) {
      case KioskState.welcome:
        return KioskWelcomeScreen(
          storeId: _storeId,
          deviceId: _deviceId,
          onStart: () {
            setState(() {
              _currentState = KioskState.auth;
            });
          },
        );

      case KioskState.auth:
        return KioskAuthScreen(
          onAuthenticated: (phone) {
            setState(() {
              _customerPhone = phone;
              _currentState = KioskState.checkout;
            });
          },
          onCancel: _resetToWelcome,
        );

      case KioskState.checkout:
        return KioskCheckoutScreen(
          storeId: _storeId,
          customerPhone: _customerPhone,
          onIdleTimeout: _resetToWelcome,
          onPaymentCompleted: (orderId, exitQr) {
            setState(() {
              _currentOrderId = orderId;
              _currentExitQr = exitQr;
              _currentState = KioskState.exitPass;
            });
          },
        );

      case KioskState.exitPass:
        return KioskExitPassScreen(
          orderId: _currentOrderId,
          exitQr: _currentExitQr,
          customerPhone: _customerPhone,
          onDone: _resetToWelcome,
        );
    }
  }
}

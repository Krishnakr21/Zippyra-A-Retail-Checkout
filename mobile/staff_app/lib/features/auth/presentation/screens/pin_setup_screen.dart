import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../bloc/auth_bloc.dart';
import '../widgets/pin_pad.dart';

class PinSetupScreen extends StatefulWidget {
  final VoidCallback onCompleted;

  const PinSetupScreen({super.key, required this.onCompleted});

  @override
  State<PinSetupScreen> createState() => _PinSetupScreenState();
}

class _PinSetupScreenState extends State<PinSetupScreen> {
  String? _firstPin;
  String _message = 'Enter a 4-digit PIN for quick login';

  @override
  Widget build(BuildContext context) {
    return BlocListener<AuthBloc, AuthState>(
      listener: (context, state) {
        if (state is AuthStepUpRequired) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Session expired for PIN setup. Please log in with OTP again.'),
              backgroundColor: Colors.orange,
            ),
          );
          context.read<AuthBloc>().add(AuthLogoutRequested());
        } else if (state is AuthError) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(state.message), backgroundColor: Colors.red),
          );
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Set Up Quick PIN'),
          actions: [
            TextButton(
              onPressed: widget.onCompleted,
              child: const Text('Set up later'),
            ),
          ],
        ),
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              children: [
                const SizedBox(height: 20),
                Text(
                  _message,
                  style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
                Expanded(
                  child: PinPadWidget(
                    maxLength: 4,
                    onCompleted: (pin) {
                      if (_firstPin == null) {
                        setState(() {
                          _firstPin = pin;
                          _message = 'Re-enter your PIN to confirm';
                        });
                      } else {
                        if (pin == _firstPin) {
                          context.read<AuthBloc>().add(AuthPinSetupRequested(pin));
                          widget.onCompleted();
                        } else {
                          setState(() {
                            _firstPin = null;
                            _message = 'PINs did not match. Enter a 4-digit PIN again';
                          });
                        }
                      }
                    },
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

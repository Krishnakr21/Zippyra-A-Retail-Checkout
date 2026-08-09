import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/store_session_bloc.dart';

class StoreBindingScreen extends StatelessWidget {
  const StoreBindingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocListener<StoreSessionBloc, StoreSessionState>(
      listener: (context, state) {
        if (state is StoreSessionActive) {
          context.go('/store/bound');
        } else if (state is StoreSessionBindFailure) {
          final failure = state.failure;
          if (failure is StoreClosedFailure) {
            context.go('/store/closed', extra: failure.message);
          } else if (failure is StoreAtCapacityFailure) {
            context.go('/store/at_capacity', extra: failure.message);
          } else if (failure is StoreGeofenceMismatchFailure) {
            context.go('/store/geofence_mismatch', extra: failure.message);
          } else {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(failure.message), backgroundColor: Colors.red),
            );
            context.go('/store/list');
          }
        }
      },
      child: Scaffold(
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: const [
                CircularProgressIndicator(strokeWidth: 3),
                SizedBox(height: 24),
                Text(
                  'Verifying Entrance & Geofence...',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                SizedBox(height: 8),
                Text(
                  'Please wait while we validate your location and check store capacity.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: ZippyraColors.textSecondary),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

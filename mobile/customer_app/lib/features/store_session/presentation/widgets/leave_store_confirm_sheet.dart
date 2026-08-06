import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/store_session_bloc.dart';

void showLeaveStoreConfirmSheet(BuildContext context) {
  showModalBottomSheet(
    context: context,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (ctx) {
      return Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Leave Store Session?',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            const Text(
              'Unbinding your session will clear your active in-store cart. If you have completed checkout, proceed to gate validation.',
              style: TextStyle(color: ZippyraColors.textSecondary),
            ),
            const SizedBox(height: 24),
            ZButton(
              label: 'Yes, Leave Store',
              onPressed: () {
                Navigator.pop(ctx);
                context.read<StoreSessionBloc>().add(const UnbindRequested(reason: 'manual_leave'));
                context.go('/store/list');
              },
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Stay in Store'),
            ),
          ],
        ),
      );
    },
  );
}

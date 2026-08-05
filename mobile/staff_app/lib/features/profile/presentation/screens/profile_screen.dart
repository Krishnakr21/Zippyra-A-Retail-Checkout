import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    String staffId = 'staff-001';
    String role = 'MANAGER';
    String storeName = 'Downtown Superstore';

    if (authState is AuthAuthenticated) {
      staffId = authState.staffId;
      role = authState.role;
      storeName = authState.storeName;
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Staff Profile')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          children: [
            const CircleAvatar(
              radius: 40,
              backgroundColor: ZippyraColors.primary,
              child: Icon(Icons.person, size: 48, color: Colors.white),
            ),
            const SizedBox(height: 16),
            Text(
              'Staff ID: $staffId',
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 4),
            Text(
              'Role: $role • $storeName',
              style: const TextStyle(color: Colors.grey, fontSize: 14),
            ),
            const SizedBox(height: 32),

            Card(
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.language),
                    title: const Text('App Language'),
                    subtitle: const Text('English / हिन्दी'),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Language preference saved')),
                      );
                    },
                  ),
                  const Divider(height: 1),
                  ListTile(
                    leading: const Icon(Icons.security),
                    title: const Text('Security & Session'),
                    subtitle: const Text('Auto-logout after 15 min idle'),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 32),

            ZButton(
              label: 'Log Out',
              onPressed: () {
                context.read<AuthBloc>().add(AuthLogoutRequested());
                context.go('/login');
              },
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/shift_bloc.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class ShiftHomeScreen extends StatelessWidget {
  const ShiftHomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    String staffName = 'Staff Member';
    String role = 'MANAGER';
    String storeName = 'Downtown Superstore';

    if (authState is AuthAuthenticated) {
      staffName = 'Staff #${authState.staffId}';
      role = authState.role;
      storeName = authState.storeName;
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(storeName),
        actions: [
          Chip(
            label: Text(role, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
            backgroundColor: ZippyraColors.primary.withOpacity(0.12),
          ),
          const SizedBox(width: 12),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Shift Duration Header Card
            BlocBuilder<ShiftBloc, ShiftState>(
              builder: (context, state) {
                final isActive = state is ShiftActive;
                final durationStr = isActive
                    ? _formatDuration((state as ShiftActive).elapsed)
                    : '00:00:00';

                return Card(
                  elevation: 2,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  staffName,
                                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                                ),
                                const Text(
                                  'Shift Status',
                                  style: TextStyle(color: Colors.grey, fontSize: 12),
                                ),
                              ],
                            ),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                              decoration: BoxDecoration(
                                color: isActive ? Colors.green[100] : Colors.grey[200],
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                isActive ? 'ACTIVE SHIFT' : 'OFF SHIFT',
                                style: TextStyle(
                                  color: isActive ? Colors.green[900] : Colors.grey[800],
                                  fontWeight: FontWeight.bold,
                                  fontSize: 11,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        Text(
                          durationStr,
                          style: const TextStyle(
                            fontSize: 36,
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                            color: ZippyraColors.primary,
                          ),
                        ),
                        const SizedBox(height: 16),
                        ZButton(
                          label: isActive ? 'End Shift' : 'Start Shift',
                          onPressed: () {
                            if (isActive) {
                              context.read<ShiftBloc>().add(ShiftEndRequested());
                            } else {
                              context.read<ShiftBloc>().add(ShiftStartRequested());
                            }
                          },
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),

            const SizedBox(height: 28),
            const Text(
              'Quick Actions',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),

            // Large Touch Target Quick Actions Grid
            GridView.count(
              crossAxisCount: 2,
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
              childAspectRatio: 1.3,
              children: [
                _buildQuickActionCard(
                  context,
                  title: 'Stock Count',
                  subtitle: 'Audit store inventory',
                  icon: Icons.inventory_2_outlined,
                  color: Colors.blue,
                  onTap: () => context.push('/home/inventory/stock-count'),
                ),
                _buildQuickActionCard(
                  context,
                  title: 'Low Stock',
                  subtitle: 'View reorder alerts',
                  icon: Icons.warning_amber_outlined,
                  color: Colors.orange[800]!,
                  onTap: () => context.push('/home/inventory/low-stock'),
                ),
                _buildQuickActionCard(
                  context,
                  title: 'Cash Payment',
                  subtitle: 'Process cashier payment',
                  icon: Icons.point_of_sale_outlined,
                  color: Colors.green,
                  onTap: () => context.push('/home/pos-assist/cash-payment'),
                ),
                _buildQuickActionCard(
                  context,
                  title: 'GRN Receive',
                  subtitle: 'Receive warehouse delivery',
                  icon: Icons.local_shipping_outlined,
                  color: Colors.purple,
                  onTap: () => context.push('/home/inventory/grn'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickActionCard(
    BuildContext context, {
    required String title,
    required String subtitle,
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: InkWell(
        borderRadius: BorderRadius.circular(16),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              CircleAvatar(
                backgroundColor: color.withOpacity(0.12),
                child: Icon(icon, color: color),
              ),
              const SizedBox(height: 12),
              Text(
                title,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              Text(
                subtitle,
                style: const TextStyle(color: Colors.grey, fontSize: 11),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDuration(Duration duration) {
    String twoDigits(int n) => n.toString().padLeft(2, '0');
    final hours = twoDigits(duration.inHours);
    final minutes = twoDigits(duration.inMinutes.remainder(60));
    final seconds = twoDigits(duration.inSeconds.remainder(60));
    return '$hours:$minutes:$seconds';
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';
import 'package:zippyra_core/widgets/zbadge.dart';
import 'presentation/bloc/shift_bloc.dart';

class ShiftDashboardScreen extends StatefulWidget {
  const ShiftDashboardScreen({super.key});

  @override
  State<ShiftDashboardScreen> createState() => _ShiftDashboardScreenState();
}

class _ShiftDashboardScreenState extends State<ShiftDashboardScreen> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    context.read<ShiftBloc>().add(ShiftCurrentRequested());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<ShiftBloc>().add(ShiftCurrentRequested());
    }
  }

  String _formatDuration(Duration duration) {
    final hours = duration.inHours;
    final minutes = duration.inMinutes.remainder(60);
    final seconds = duration.inSeconds.remainder(60);
    if (hours > 0) {
      return '${hours}h ${minutes}m ${seconds}s';
    }
    return '${minutes}m ${seconds}s';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Shift Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.person_outline),
            onPressed: () => context.push('/profile'),
          ),
        ],
      ),
      body: BlocBuilder<ShiftBloc, ShiftState>(
        builder: (context, state) {
          final isActive = state is ShiftActive;
          final durationText = isActive ? _formatDuration(state.elapsed) : 'Inactive';

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: isActive
                        ? [ZippyraColors.primaryBlue, const Color(0xFF0D3E75)]
                        : [Colors.grey.shade700, Colors.grey.shade900],
                  ),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            ZBadge(
                              text: isActive ? 'ACTIVE SHIFT' : 'NO SHIFT',
                              color: isActive ? ZippyraColors.successGreen : Colors.orange,
                            ),
                            const SizedBox(height: 8),
                            const Text(
                              'Store #IN-104 • Indiranagar',
                              style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              'Shift Duration: $durationText',
                              style: const TextStyle(color: Colors.white70),
                            ),
                          ],
                        ),
                        const Icon(Icons.access_time_filled, color: Colors.white, size: 40),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        if (isActive)
                          ElevatedButton.icon(
                            key: const Key('end_shift_button'),
                            onPressed: () => context.read<ShiftBloc>().add(ShiftEndRequested()),
                            icon: const Icon(Icons.stop_circle_outlined, color: Colors.red),
                            label: const Text('End Shift', style: TextStyle(color: Colors.red)),
                            style: ElevatedButton.styleFrom(backgroundColor: Colors.white),
                          )
                        else
                          ElevatedButton.icon(
                            key: const Key('start_shift_button'),
                            onPressed: () => context.read<ShiftBloc>().add(ShiftStartRequested()),
                            icon: const Icon(Icons.play_circle_fill, color: Colors.green),
                            label: const Text('Start Shift', style: TextStyle(color: Colors.green)),
                            style: ElevatedButton.styleFrom(backgroundColor: Colors.white),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              const Text('Quick Action Modules', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
              const SizedBox(height: 12),
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                children: [
                  _buildTile(context, 'Stock Count', Icons.inventory_2_outlined, ZippyraColors.primaryBlue, '/inventory/stock_count', badgeText: '2 Pending'),
                  _buildTile(context, 'Receive GRN', Icons.local_shipping_outlined, ZippyraColors.primaryBlue, '/inventory/grn'),
                  _buildTile(context, 'RFID Deactivate', Icons.nfc, ZippyraColors.accentOrange, '/devices/rfid_deactivation'),
                  _buildTile(context, 'Device Alerts', Icons.warning_amber_rounded, ZippyraColors.errorRed, '/devices/device_alerts', badgeText: '1 Offline'),
                  _buildTile(context, 'Price Check', Icons.price_check, ZippyraColors.successGreen, '/pos_assist/price_check'),
                  _buildTile(context, 'Cash Payment', Icons.payments_outlined, ZippyraColors.successGreen, '/pos_assist/cash_payment'),
                  _buildTile(context, 'Customer Assist', Icons.support_agent, ZippyraColors.primaryBlue, '/customer_assist'),
                ],
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildTile(BuildContext context, String title, IconData icon, Color color, String route, {String? badgeText}) {
    return ZCard(
      onTap: () => context.push(route),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 36, color: color),
          const SizedBox(height: 8),
          Text(title, textAlign: TextAlign.center, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
          if (badgeText != null) ...[
            const SizedBox(height: 4),
            ZBadge(text: badgeText, color: color),
          ],
        ],
      ),
    );
  }
}

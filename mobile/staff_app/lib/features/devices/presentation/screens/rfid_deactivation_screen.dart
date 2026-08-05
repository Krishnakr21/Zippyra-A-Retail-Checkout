import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/devices_bloc.dart';
import '../bloc/devices_event.dart';
import '../bloc/devices_state.dart';
import '../../domain/models/exit_attempt_model.dart';

class RfidDeactivationScreen extends StatefulWidget {
  final String storeId;

  const RfidDeactivationScreen({Key? key, required this.storeId}) : super(key: key);

  @override
  State<RfidDeactivationScreen> createState() => _RfidDeactivationScreenState();
}

class _RfidDeactivationScreenState extends State<RfidDeactivationScreen> {
  @override
  void initState() {
    super.initState();
    context.read<DevicesBloc>().add(RecentExitAttemptsRequested(widget.storeId));
  }

  void _showStaffOverrideDialog(ExitAttemptModel attempt) {
    String selectedReason = 'CUSTOMER_VERIFIED_RECEIPT';

    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: Text('Staff Override for Gate ${attempt.gateId}'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Order ID: ${attempt.orderId}'),
            const SizedBox(height: 8),
            Text('Result: ${attempt.result}', style: const TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            const Text('Select Override Reason:', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            DropdownButtonFormField<String>(
              value: selectedReason,
              items: const [
                DropdownMenuItem(value: 'CUSTOMER_VERIFIED_RECEIPT', child: Text('Customer Verified Receipt')),
                DropdownMenuItem(value: 'HARDWARE_MALFUNCTION', child: Text('Hardware Malfunction')),
                DropdownMenuItem(value: 'SYSTEM_BYPASS', child: Text('System Bypass')),
              ],
              onChanged: (val) {
                if (val != null) selectedReason = val;
              },
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogCtx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              context.read<DevicesBloc>().add(
                    StaffOverrideRequested(
                      orderId: attempt.orderId,
                      gateId: attempt.gateId,
                      reason: selectedReason,
                      storeId: widget.storeId,
                    ),
                  );
              Navigator.pop(dialogCtx);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.indigo, foregroundColor: Colors.white),
            child: const Text('Execute Override'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Gate Exit & RFID Audit Monitor'),
      ),
      body: BlocListener<DevicesBloc, DevicesState>(
        listener: (context, state) {
          if (state is StaffOverrideSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Staff Override executed successfully for Order ${state.orderId}'),
                backgroundColor: Colors.green,
              ),
            );
            context.read<DevicesBloc>().add(RecentExitAttemptsRequested(widget.storeId));
          }
        },
        child: BlocBuilder<DevicesBloc, DevicesState>(
          builder: (context, state) {
            if (state is DevicesLoading) {
              return const Center(child: CircularProgressIndicator());
            }
            if (state is DevicesError) {
              return Center(child: Text(state.message));
            }
            if (state is RecentExitAttemptsLoaded) {
              if (state.attempts.isEmpty) {
                return const Center(child: Text('No recent gate exit attempts for this store.'));
              }
              return ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: state.attempts.length,
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemBuilder: (context, index) {
                  final attempt = state.attempts[index];
                  return Card(
                    elevation: 2,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: ListTile(
                      leading: CircleAvatar(
                        backgroundColor: attempt.isAlarm ? Colors.red.shade100 : Colors.indigo.shade100,
                        child: Icon(
                          attempt.isAlarm ? Icons.alarm : Icons.shield_outlined,
                          color: attempt.isAlarm ? Colors.red : Colors.indigo,
                        ),
                      ),
                      title: Text('Order: ${attempt.orderId} (Gate: ${attempt.gateId})',
                          style: const TextStyle(fontWeight: FontWeight.bold)),
                      subtitle: Text('Status: ${attempt.result}'),
                      trailing: ElevatedButton(
                        onPressed: () => _showStaffOverrideDialog(attempt),
                        style: ElevatedButton.styleFrom(backgroundColor: Colors.orange, foregroundColor: Colors.white),
                        child: const Text('Staff Override'),
                      ),
                    ),
                  );
                },
              );
            }
            return const SizedBox.shrink();
          },
        ),
      ),
    );
  }
}

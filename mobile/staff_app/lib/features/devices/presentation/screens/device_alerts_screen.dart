import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/devices_bloc.dart';
import '../bloc/devices_event.dart';
import '../bloc/devices_state.dart';

class DeviceAlertsScreen extends StatefulWidget {
  final String storeId;

  const DeviceAlertsScreen({Key? key, required this.storeId}) : super(key: key);

  @override
  State<DeviceAlertsScreen> createState() => _DeviceAlertsScreenState();
}

class _DeviceAlertsScreenState extends State<DeviceAlertsScreen> {
  @override
  void initState() {
    super.initState();
    context.read<DevicesBloc>().add(DeviceAlertsRequested(widget.storeId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Unresolved Hardware Alerts'),
      ),
      body: BlocBuilder<DevicesBloc, DevicesState>(
        builder: (context, state) {
          if (state is DevicesLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is DevicesError) {
            return Center(child: Text(state.message));
          }
          if (state is DeviceAlertsLoaded) {
            final unresolved = state.alerts.where((a) => a.resolvedAt == null).toList();
            if (unresolved.isEmpty) {
              return const Center(child: Text('All hardware alerts resolved.'));
            }
            return ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: unresolved.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final alert = unresolved[index];
                return Card(
                  elevation: 2,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: ListTile(
                    leading: const CircleAvatar(
                      backgroundColor: Colors.redAccent,
                      child: Icon(Icons.warning, color: Colors.white),
                    ),
                    title: Text(
                      alert.alertType,
                      style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.red),
                    ),
                    subtitle: Text(
                      alert.detail?['label']?.toString() ?? 'Device ID: ${alert.deviceId}',
                    ),
                    trailing: ElevatedButton(
                      key: Key('resolve_btn_${alert.id}'),
                      onPressed: () {
                        context.read<DevicesBloc>().add(DeviceAlertResolveRequested(alert.id));
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.green,
                        foregroundColor: Colors.white,
                      ),
                      child: const Text('Resolve'),
                    ),
                  ),
                );
              },
            );
          }
          return const SizedBox.shrink();
        },
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/devices_bloc.dart';
import '../bloc/devices_event.dart';
import '../bloc/devices_state.dart';

class DeviceListScreen extends StatefulWidget {
  final String storeId;

  const DeviceListScreen({Key? key, required this.storeId}) : super(key: key);

  @override
  State<DeviceListScreen> createState() => _DeviceListScreenState();
}

class _DeviceListScreenState extends State<DeviceListScreen> {
  @override
  void initState() {
    super.initState();
    context.read<DevicesBloc>().add(DeviceListRequested(widget.storeId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Store Hardware Devices'),
      ),
      body: BlocBuilder<DevicesBloc, DevicesState>(
        builder: (context, state) {
          if (state is DevicesLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is DevicesError) {
            return Center(child: Text(state.message));
          }
          if (state is DeviceListLoaded) {
            if (state.devices.isEmpty) {
              return const Center(child: Text('No devices registered at this store.'));
            }
            return ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: state.devices.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final d = state.devices[index];
                return Card(
                  elevation: 2,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: Colors.indigo.shade100,
                      child: Icon(
                        d.deviceType == 'GATE'
                            ? Icons.door_sliding
                            : d.deviceType == 'RFID_PAD'
                                ? Icons.nfc
                                : Icons.qr_code_scanner,
                        color: Colors.indigo,
                      ),
                    ),
                    title: Text(d.label, style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text('Type: ${d.deviceType} ${d.gateId != null ? "| Gate: ${d.gateId}" : ""}'),
                    trailing: Chip(
                      label: Text(
                        d.status,
                        style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                      ),
                      backgroundColor: d.status == 'ACTIVE'
                          ? Colors.green
                          : d.status == 'OFFLINE'
                              ? Colors.red
                              : Colors.amber,
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

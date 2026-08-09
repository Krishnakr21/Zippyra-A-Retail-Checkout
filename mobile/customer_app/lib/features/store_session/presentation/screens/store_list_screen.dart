import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/store_session_bloc.dart';
import '../widgets/leave_store_confirm_sheet.dart';

class StoreListScreen extends StatefulWidget {
  const StoreListScreen({super.key});

  @override
  State<StoreListScreen> createState() => _StoreListScreenState();
}

class _StoreListScreenState extends State<StoreListScreen> {
  @override
  void initState() {
    super.initState();
    // Default fetch for Bangalore coordinates in dev
    context.read<StoreSessionBloc>().add(const NearbyStoresRequested(lat: 12.9716, lng: 77.5946));
  }

  Future<void> _onEnterStorePressed() async {
    final cameraStatus = await Permission.camera.request();
    final locationStatus = await Permission.locationWhenInUse.request();

    if (cameraStatus.isGranted && locationStatus.isGranted) {
      if (mounted) context.push('/store/scan');
    } else {
      if (mounted) _showPermissionRationaleDialog();
    }
  }

  void _showPermissionRationaleDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Permissions Required'),
        content: const Text(
          'Camera and Location access are required to scan entrance QR codes and verify your store geofence location.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              openAppSettings();
            },
            child: const Text('Open Settings'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Stores Nearby'),
        actions: [
          BlocBuilder<StoreSessionBloc, StoreSessionState>(
            builder: (context, state) {
              if (state is StoreSessionActive) {
                return IconButton(
                  icon: const Icon(Icons.exit_to_app, color: Colors.red),
                  tooltip: 'Leave Store',
                  onPressed: () => showLeaveStoreConfirmSheet(context),
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: BlocConsumer<StoreSessionBloc, StoreSessionState>(
        listener: (context, state) {
          if (state is StoreSessionActive) {
            context.go('/store/bound');
          } else if (state is StoreSessionBindFailure) {
            final failure = state.failure;
            if (failure is StoreClosedFailure) {
              context.push('/store/closed', extra: failure.message);
            } else if (failure is StoreAtCapacityFailure) {
              context.push('/store/at_capacity', extra: failure.message);
            } else if (failure is StoreGeofenceMismatchFailure) {
              context.push('/store/geofence_mismatch', extra: failure.message);
            } else {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text(failure.message), backgroundColor: Colors.red),
              );
            }
          }
        },
        builder: (context, state) {
          if (state is StoreSessionRestoring) {
            return const Center(child: CircularProgressIndicator());
          }

          final nearbyStores = (state is StoreSessionNone)
              ? state.nearbyStores
              : (state is StoreSessionActive)
                  ? state.nearbyStores
                  : [];

          return Column(
            children: [
              // Active Session Banner if currently bound
              if (state is StoreSessionActive)
                Container(
                  color: Colors.blue.shade50,
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      const Icon(Icons.store, color: ZippyraColors.primaryBlue),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Active Session: ${state.session.storeName}',
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                            ),
                            const Text('You are currently inside this store', style: TextStyle(color: Colors.grey)),
                          ],
                        ),
                      ),
                      ElevatedButton(
                        onPressed: () => showLeaveStoreConfirmSheet(context),
                        style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
                        child: const Text('Leave'),
                      ),
                    ],
                  ),
                ),

              Expanded(
                child: nearbyStores.isEmpty
                    ? const Center(child: Text('No stores found nearby.'))
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: nearbyStores.length,
                        itemBuilder: (context, index) {
                          final store = nearbyStores[index];
                          return Card(
                            margin: const EdgeInsets.only(bottom: 12),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor: store.isOpen ? Colors.green.shade100 : Colors.red.shade100,
                                child: Icon(
                                  Icons.storefront,
                                  color: store.isOpen ? Colors.green : Colors.red,
                                ),
                              ),
                              title: Text(store.name, style: const TextStyle(fontWeight: FontWeight.bold)),
                              subtitle: Text('${store.address}\n${store.distanceKM} km away'),
                              trailing: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                crossAxisAlignment: CrossAxisAlignment.end,
                                children: [
                                  Text(
                                    store.isOpen ? 'OPEN' : 'CLOSED',
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                      color: store.isOpen ? Colors.green : Colors.red,
                                    ),
                                  ),
                                  Text('Cap: ${store.capacityPct}%', style: const TextStyle(fontSize: 12, color: Colors.grey)),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
              ),

              Padding(
                padding: const EdgeInsets.all(16.0),
                child: ZButton(
                  label: 'Enter Store / Scan Entrance QR',
                  onPressed: _onEnterStorePressed,
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

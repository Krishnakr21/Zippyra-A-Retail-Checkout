import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'bloc/device_sessions_bloc.dart';
import '../domain/entities/device_session.dart';

class DeviceSessionsScreen extends StatefulWidget {
  const DeviceSessionsScreen({super.key});

  @override
  State<DeviceSessionsScreen> createState() => _DeviceSessionsScreenState();
}

class _DeviceSessionsScreenState extends State<DeviceSessionsScreen> {
  @override
  void initState() {
    super.initState();
    context.read<DeviceSessionsBloc>().add(LoadDeviceSessions());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Logged-in Devices'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: BlocBuilder<DeviceSessionsBloc, DeviceSessionsState>(
        builder: (context, state) {
          if (state is DeviceSessionsLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is DeviceSessionsError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    state.message,
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: Colors.redAccent),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      context.read<DeviceSessionsBloc>().add(LoadDeviceSessions());
                    },
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (state is DeviceSessionsLoaded) {
            final sessions = state.sessions;
            final hasOtherDevices = sessions.any((s) => !s.isCurrent);

            return RefreshIndicator(
              onRefresh: () async {
                context.read<DeviceSessionsBloc>().add(LoadDeviceSessions());
              },
              child: ListView(
                padding: const EdgeInsets.all(16.0),
                children: [
                  if (hasOtherDevices) ...[
                    Card(
                      color: Colors.red.withOpacity(0.08),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(color: Colors.red.withOpacity(0.3)),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Row(
                          children: [
                            const Icon(Icons.security, color: Colors.redAccent),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: const [
                                  Text(
                                    'Active on Multiple Devices',
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                      fontSize: 15,
                                    ),
                                  ),
                                  SizedBox(height: 4),
                                  Text(
                                    'Revoke sessions if you do not recognize a device.',
                                    style: TextStyle(fontSize: 12, color: Colors.grey),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    OutlinedButton.icon(
                      onPressed: state.isRevokingAll
                          ? null
                          : () => _confirmSignOutAllOtherDevices(context),
                      icon: state.isRevokingAll
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.logout, color: Colors.redAccent),
                      label: const Text(
                        'Sign Out All Other Devices',
                        style: TextStyle(color: Colors.redAccent, fontWeight: FontWeight.bold),
                      ),
                      style: OutlinedButton.styleFrom(
                        side: const BorderSide(color: Colors.redAccent),
                        padding: const EdgeInsets.symmetric(vertical: 14),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                  const Text(
                    'ACTIVE SESSIONS',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 1.2,
                      color: Colors.grey,
                    ),
                  ),
                  const SizedBox(height: 8),
                  ...sessions.map((session) => _buildSessionTile(context, session, state)),
                ],
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildSessionTile(
    BuildContext context,
    DeviceSession session,
    DeviceSessionsLoaded state,
  ) {
    final isRevoking = state.revokingSessionId == session.id;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: session.isCurrent
            ? BorderSide(color: Theme.of(context).primaryColor, width: 1.5)
            : BorderSide.none,
      ),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: session.isCurrent
              ? Theme.of(context).primaryColor.withOpacity(0.15)
              : Colors.grey.withOpacity(0.15),
          child: Icon(
            _getDeviceIcon(session.deviceLabel),
            color: session.isCurrent
                ? Theme.of(context).primaryColor
                : Colors.grey,
          ),
        ),
        title: Row(
          children: [
            Expanded(
              child: Text(
                session.deviceLabel,
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
            ),
            if (session.isCurrent)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Theme.of(context).primaryColor.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  'This Device',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Theme.of(context).primaryColor,
                  ),
                ),
              ),
          ],
        ),
        subtitle: Text(
          session.lastUsedAt != null
              ? 'Last active: ${_formatDate(session.lastUsedAt!)}'
              : 'Created: ${_formatDate(session.createdAt)}',
          style: const TextStyle(fontSize: 12, color: Colors.grey),
        ),
        trailing: session.isCurrent
            ? null
            : isRevoking
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : IconButton(
                    icon: const Icon(Icons.close, color: Colors.redAccent),
                    onPressed: () => _confirmSignOutDevice(context, session),
                    tooltip: 'Sign out device',
                  ),
      ),
    );
  }

  IconData _getDeviceIcon(String label) {
    final lower = label.toLowerCase();
    if (lower.contains('ipad') || lower.contains('tab')) {
      return Icons.tablet_mac;
    } else if (lower.contains('mac') || lower.contains('windows') || lower.contains('desktop')) {
      return Icons.laptop;
    }
    return Icons.phone_iphone;
  }

  String _formatDate(DateTime dt) {
    return '${dt.day}/${dt.month}/${dt.year} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }

  void _confirmSignOutDevice(BuildContext context, DeviceSession session) {
    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Sign Out Device?'),
        content: Text('Are you sure you want to sign out "${session.deviceLabel}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogCtx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent),
            onPressed: () {
              Navigator.pop(dialogCtx);
              context.read<DeviceSessionsBloc>().add(RevokeDeviceSession(session.id));
            },
            child: const Text('Sign Out'),
          ),
        ],
      ),
    );
  }

  void _confirmSignOutAllOtherDevices(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Sign Out All Other Devices?'),
        content: const Text('This will log you out everywhere except this device.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogCtx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent),
            onPressed: () {
              Navigator.pop(dialogCtx);
              context.read<DeviceSessionsBloc>().add(RevokeAllOtherDeviceSessions());
            },
            child: const Text('Sign Out All'),
          ),
        ],
      ),
    );
  }
}

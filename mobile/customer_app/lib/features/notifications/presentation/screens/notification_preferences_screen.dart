import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../bloc/notifications_bloc.dart';
import '../bloc/notifications_event.dart';
import '../bloc/notifications_state.dart';
import '../../domain/entities/notification_preference.dart';

class NotificationPreferencesScreen extends StatefulWidget {
  const NotificationPreferencesScreen({Key? key}) : super(key: key);

  @override
  State<NotificationPreferencesScreen> createState() => _NotificationPreferencesScreenState();
}

class _NotificationPreferencesScreenState extends State<NotificationPreferencesScreen> {
  @override
  void initState() {
    super.initState();
    context.read<NotificationsBloc>().add(PreferencesRequested());
  }

  String _getLabel(String notifType) {
    switch (notifType) {
      case 'ORDER_UPDATES':
        return 'Order Updates';
      case 'LOYALTY_UPDATES':
        return 'Loyalty & Rewards';
      case 'MARKETING':
        return 'Promotions & Offers';
      default:
        return notifType.replaceAll('_', ' ');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Notification Preferences'),
      ),
      body: BlocConsumer<NotificationsBloc, NotificationsState>(
        listener: (context, state) {
          if (state.errorMessage != null && state.errorMessage!.isNotEmpty) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(state.errorMessage!),
                backgroundColor: Colors.red,
              ),
            );
          }
        },
        builder: (context, state) {
          if (state.isPreferencesLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          // Known types array for predictable UI list
          const knownTypes = ['ORDER_UPDATES', 'LOYALTY_UPDATES', 'MARKETING'];
          final prefMap = {for (var p in state.preferences) p.notificationType: p};

          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: knownTypes.length,
            separatorBuilder: (_, __) => const Divider(height: 24),
            itemBuilder: (context, index) {
              final type = knownTypes[index];
              final pref = prefMap[type] ??
                  NotificationPreference(
                    notificationType: type,
                    channel: 'BOTH',
                    isMandatory: type == 'ORDER_UPDATES',
                  );

              final isMandatory = pref.isMandatory || type == 'ORDER_UPDATES';

              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _getLabel(type),
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            if (isMandatory)
                              const Padding(
                                padding: EdgeInsets.only(top: 4),
                                child: Text(
                                  'Mandatory transactional update. Cannot be disabled.',
                                  key: Key('mandatory_note_text'),
                                  style: TextStyle(fontSize: 12, color: Colors.grey),
                                ),
                              ),
                          ],
                        ),
                      ),
                      Switch(
                        key: Key('switch_$type'),
                        value: isMandatory ? true : pref.isEnabled,
                        onChanged: isMandatory
                            ? null // Non-interactive disabled toggle for mandatory notifications
                            : (val) {
                                context.read<NotificationsBloc>().add(
                                      PreferenceToggled(
                                        notificationType: type,
                                        enabled: val,
                                      ),
                                    );
                              },
                      ),
                    ],
                  ),
                ],
              );
            },
          );
        },
      ),
    );
  }
}

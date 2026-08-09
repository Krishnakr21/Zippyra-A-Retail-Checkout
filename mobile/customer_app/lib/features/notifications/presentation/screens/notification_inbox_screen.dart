import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../bloc/notifications_bloc.dart';
import '../bloc/notifications_event.dart';
import '../bloc/notifications_state.dart';
import '../../../../core/services/deep_link_service.dart';
import '../../../../injection_container.dart';

class NotificationInboxScreen extends StatefulWidget {
  const NotificationInboxScreen({Key? key}) : super(key: key);

  @override
  State<NotificationInboxScreen> createState() => _NotificationInboxScreenState();
}

class _NotificationInboxScreenState extends State<NotificationInboxScreen> {
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    context.read<NotificationsBloc>().add(const InboxRequested(refresh: true));
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      context.read<NotificationsBloc>().add(InboxNextPageRequested());
    }
  }

  String _formatRelativeTime(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return 'Just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    return '${diff.inDays}d ago';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
      ),
      body: BlocBuilder<NotificationsBloc, NotificationsState>(
        builder: (context, state) {
          if (state.isInboxLoading && state.inbox.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state.inbox.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: const [
                  Icon(Icons.notifications_off_outlined, size: 64, color: Colors.grey),
                  SizedBox(height: 16),
                  Text(
                    'No notifications yet',
                    style: TextStyle(fontSize: 16, color: Colors.grey, fontWeight: FontWeight.w500),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async {
              context.read<NotificationsBloc>().add(const InboxRequested(refresh: true));
            },
            child: ListView.separated(
              controller: _scrollController,
              padding: const EdgeInsets.all(12),
              itemCount: state.inbox.length + (state.hasReachedMaxInbox ? 0 : 1),
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                if (index >= state.inbox.length) {
                  return const Padding(
                    padding: EdgeInsets.all(16.0),
                    child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
                  );
                }

                final item = state.inbox[index];

                return ListTile(
                  key: Key('inbox_item_${item.id}'),
                  onTap: () {
                    context.read<NotificationsBloc>().add(NotificationMarkedRead(item.id));
                    if (item.deepLink != null && item.deepLink!.isNotEmpty) {
                      sl<DeepLinkService>().handleDeepLink(item.deepLink);
                    }
                  },
                  leading: Stack(
                    alignment: Alignment.topRight,
                    children: [
                      CircleAvatar(
                        backgroundColor: Colors.indigo.shade50,
                        child: Icon(
                          item.notificationType == 'ORDER_UPDATES'
                              ? Icons.shopping_bag_outlined
                              : item.notificationType == 'LOYALTY_UPDATES'
                                  ? Icons.stars_outlined
                                  : Icons.notifications_none_outlined,
                          color: Colors.indigo,
                        ),
                      ),
                      if (!item.isRead)
                        Container(
                          width: 10,
                          height: 10,
                          decoration: const BoxDecoration(
                            color: Colors.blue,
                            shape: BoxShape.circle,
                          ),
                        ),
                    ],
                  ),
                  title: Text(
                    item.title,
                    style: TextStyle(
                      fontWeight: item.isRead ? FontWeight.normal : FontWeight.bold,
                      fontSize: 15,
                    ),
                  ),
                  subtitle: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 4),
                      Text(
                        item.body,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontSize: 13, color: Colors.black87),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _formatRelativeTime(item.createdAt),
                        style: const TextStyle(fontSize: 11, color: Colors.grey),
                      ),
                    ],
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }
}

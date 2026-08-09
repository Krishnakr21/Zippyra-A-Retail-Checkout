import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class NotificationsScreen extends StatefulWidget {
  const NotificationsScreen({super.key});

  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  final List<Map<String, dynamic>> _todayNotifications = [
    {
      'id': 'notif-1',
      'icon': '✅',
      'title': 'Order ZPY-8821 completed — exit validated!',
      'time': '9:43 AM',
      'isUnread': true,
      'dotColor': Color(0xFF2563EB),
      'iconBg': Color(0xFFDCFCE7),
    },
    {
      'id': 'notif-2',
      'icon': '⭐',
      'title': '+52 Zippy Points! Now 392 total',
      'time': '9:43 AM',
      'isUnread': true,
      'dotColor': Color(0xFFF59E0B),
      'iconBg': Color(0xFFFEF3C7),
    },
    {
      'id': 'notif-3',
      'icon': '🎟️',
      'title': '20% off Amul products — today only!',
      'time': '8:00 AM',
      'isUnread': false,
      'dotColor': Colors.transparent,
      'iconBg': Color(0xFFFEE2E2),
    },
  ];

  final List<Map<String, dynamic>> _yesterdayNotifications = [
    {
      'id': 'notif-4',
      'icon': '💬',
      'title': 'WhatsApp receipt delivered for ZPY-7134',
      'time': '6:24 PM',
      'isUnread': false,
      'dotColor': Colors.transparent,
      'iconBg': Color(0xFFF1F5F9),
    },
    {
      'id': 'notif-5',
      'icon': '🔄',
      'title': 'Return approved for Dove Shampoo',
      'time': '3:10 PM',
      'isUnread': false,
      'dotColor': Colors.transparent,
      'iconBg': Color(0xFFEFF6FF),
    },
  ];

  void _markAllRead() {
    setState(() {
      for (final item in _todayNotifications) {
        item['isUnread'] = false;
      }
      for (final item in _yesterdayNotifications) {
        item['isUnread'] = false;
      }
    });
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('All notifications marked as read ✓'),
        duration: Duration(seconds: 1),
      ),
    );
  }

  void _removeNotification(String listType, int index, String title) {
    setState(() {
      if (listType == 'today') {
        _todayNotifications.removeAt(index);
      } else {
        _yesterdayNotifications.removeAt(index);
      }
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Deleted: "$title"'),
        action: SnackBarAction(
          label: 'DISMISS',
          textColor: Colors.white,
          onPressed: () {},
        ),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final hasNoNotifications = _todayNotifications.isEmpty && _yesterdayNotifications.isEmpty;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        iconTheme: const IconThemeData(color: Color(0xFF1E293B)),
        title: const Text(
          'Notifications',
          style: TextStyle(fontWeight: FontWeight.w900, fontSize: 18, color: Color(0xFF1E293B)),
        ),
        actions: [
          if (!hasNoNotifications)
            TextButton(
              onPressed: _markAllRead,
              child: const Text(
                'Mark All Read',
                style: TextStyle(color: Color(0xFF2563EB), fontWeight: FontWeight.bold, fontSize: 13),
              ),
            ),
          const SizedBox(width: 8),
        ],
      ),
      body: SafeArea(
        child: hasNoNotifications
            ? Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.notifications_off_outlined, size: 64, color: Color(0xFF94A3B8)),
                    const SizedBox(height: 12),
                    const Text(
                      'No Notifications Right Now',
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Color(0xFF1E293B)),
                    ),
                    const SizedBox(height: 6),
                    const Text(
                      'You are all caught up!',
                      style: TextStyle(color: Color(0xFF64748B), fontSize: 12),
                    ),
                  ],
                ),
              )
            : ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  if (_todayNotifications.isNotEmpty) ...[
                    const Text(
                      'TODAY',
                      style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: Color(0xFF64748B), letterSpacing: 0.5),
                    ),
                    const SizedBox(height: 8),
                    ..._todayNotifications.asMap().entries.map((entry) {
                      final idx = entry.key;
                      final item = entry.value;
                      return _buildDismissibleNotificationCard('today', idx, item);
                    }),
                    const SizedBox(height: 20),
                  ],
                  if (_yesterdayNotifications.isNotEmpty) ...[
                    const Text(
                      'YESTERDAY',
                      style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: Color(0xFF64748B), letterSpacing: 0.5),
                    ),
                    const SizedBox(height: 8),
                    ..._yesterdayNotifications.asMap().entries.map((entry) {
                      final idx = entry.key;
                      final item = entry.value;
                      return _buildDismissibleNotificationCard('yesterday', idx, item);
                    }),
                  ],
                  const SizedBox(height: 20),
                  const Center(
                    child: Text(
                      '💡 Swipe left or right on any notification to delete',
                      style: TextStyle(fontSize: 11, color: Color(0xFF94A3B8), fontWeight: FontWeight.w500),
                    ),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _buildDismissibleNotificationCard(String listType, int index, Map<String, dynamic> item) {
    final id = item['id'] as String;
    final title = item['title'] as String;
    final time = item['time'] as String;
    final icon = item['icon'] as String;
    final isUnread = item['isUnread'] as bool;
    final dotColor = item['dotColor'] as Color;
    final iconBg = item['iconBg'] as Color;

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      child: Dismissible(
        key: Key(id),
        background: Container(
          alignment: Alignment.centerLeft,
          padding: const EdgeInsets.only(left: 20),
          decoration: BoxDecoration(
            color: const Color(0xFFEF4444),
            borderRadius: BorderRadius.circular(16),
          ),
          child: const Icon(Icons.delete_outline, color: Colors.white, size: 26),
        ),
        secondaryBackground: Container(
          alignment: Alignment.centerRight,
          padding: const EdgeInsets.only(right: 20),
          decoration: BoxDecoration(
            color: const Color(0xFFEF4444),
            borderRadius: BorderRadius.circular(16),
          ),
          child: const Icon(Icons.delete_outline, color: Colors.white, size: 26),
        ),
        onDismissed: (_) => _removeNotification(listType, index, title),
        child: Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: isUnread ? const Color(0xFFBFDBFE) : const Color(0xFFE2E8F0),
              width: isUnread ? 1.5 : 1,
            ),
            boxShadow: const [
              BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2)),
            ],
          ),
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: iconBg,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Center(
                  child: Text(icon, style: const TextStyle(fontSize: 22)),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: TextStyle(
                        fontWeight: isUnread ? FontWeight.w900 : FontWeight.bold,
                        fontSize: 13,
                        color: const Color(0xFF1E293B),
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      time,
                      style: const TextStyle(fontSize: 11, color: Color(0xFF94A3B8), fontWeight: FontWeight.w500),
                    ),
                  ],
                ),
              ),
              if (isUnread)
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: dotColor,
                    shape: BoxShape.circle,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

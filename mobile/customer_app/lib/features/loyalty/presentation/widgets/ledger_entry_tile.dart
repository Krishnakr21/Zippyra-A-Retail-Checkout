import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../domain/entities/loyalty_ledger_entry.dart';

class LedgerEntryTile extends StatelessWidget {
  final LoyaltyLedgerEntry entry;

  const LedgerEntryTile({super.key, required this.entry});

  @override
  Widget build(BuildContext context) {
    IconData icon;
    Color iconColor;
    String sign = '';

    final type = entry.entryType.toUpperCase();
    if (type.contains('EARN')) {
      icon = Icons.arrow_upward;
      iconColor = Colors.green;
      sign = '+';
    } else if (type.contains('REVERSAL')) {
      icon = Icons.replay;
      iconColor = Colors.blue;
      sign = '';
    } else {
      icon = Icons.arrow_downward;
      iconColor = Colors.orange[800]!;
      sign = '';
    }

    final formattedDate = '${entry.createdAt.day}/${entry.createdAt.month}/${entry.createdAt.year}';
    final isOrderRef = entry.referenceType == 'ORDER' && entry.referenceId != null && entry.referenceId!.isNotEmpty;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: iconColor.withOpacity(0.12),
          child: Icon(icon, color: iconColor, size: 22),
        ),
        title: Row(
          children: [
            Text(
              _formatEntryTypeTitle(entry.entryType),
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
            ),
            if (isOrderRef) ...[
              const SizedBox(width: 8),
              InkWell(
                onTap: () {
                  context.push('/order/${entry.referenceId!}');
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.blue[50],
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    'Order #${entry.referenceId!.substring(0, entry.referenceId!.length > 8 ? 8 : entry.referenceId!.length)}',
                    style: const TextStyle(fontSize: 11, color: Colors.blue, fontWeight: FontWeight.w600),
                  ),
                ),
              ),
            ],
          ],
        ),
        subtitle: Text(
          '$formattedDate • Balance: ${entry.balanceAfter} pts',
          style: const TextStyle(fontSize: 12, color: Colors.grey),
        ),
        trailing: Text(
          '$sign${entry.pointsDelta} pts',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 15,
            color: entry.pointsDelta > 0 ? Colors.green : (entry.pointsDelta < 0 ? Colors.red : Colors.grey),
          ),
        ),
      ),
    );
  }

  String _formatEntryTypeTitle(String type) {
    switch (type.toUpperCase()) {
      case 'EARN':
        return 'Points Earned';
      case 'REDEEM_RESERVE':
      case 'REDEEM_COMMIT':
        return 'Points Redeemed';
      case 'REDEEM_RELEASE':
        return 'Points Restored';
      case 'REVERSAL':
        return 'Points Reversed';
      default:
        return 'Points Adjustment';
    }
  }
}

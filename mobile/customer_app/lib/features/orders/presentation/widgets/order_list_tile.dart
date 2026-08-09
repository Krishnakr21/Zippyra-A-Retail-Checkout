import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/order_summary.dart';
import 'order_status_chip.dart';

class OrderListTile extends StatelessWidget {
  final OrderSummary order;
  final VoidCallback onTap;

  const OrderListTile({
    super.key,
    required this.order,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final dateStr = '${order.createdAt.day}/${order.createdAt.month}/${order.createdAt.year}';

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        onTap: onTap,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        title: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Text(
                order.storeName,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            OrderStatusChip(status: order.status),
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 6),
            Text(
              '$dateStr • ${order.itemCount} items',
              style: const TextStyle(color: Colors.grey, fontSize: 13),
            ),
            const SizedBox(height: 4),
            Text(
              CurrencyFormatter.formatPaise(order.totalPaise),
              style: const TextStyle(fontWeight: FontWeight.bold, color: ZippyraColors.primary, fontSize: 15),
            ),
          ],
        ),
        trailing: const Icon(Icons.chevron_right, color: Colors.grey),
      ),
    );
  }
}

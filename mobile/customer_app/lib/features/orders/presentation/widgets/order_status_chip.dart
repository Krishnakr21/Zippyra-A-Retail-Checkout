import 'package:flutter/material.dart';

class OrderStatusChip extends StatelessWidget {
  final String status;

  const OrderStatusChip({
    super.key,
    required this.status,
  });

  @override
  Widget build(BuildContext context) {
    Color bg;
    Color fg;
    String label;

    switch (status) {
      case 'COMPLETED':
        bg = Colors.green[50]!;
        fg = Colors.green[800]!;
        label = 'Completed';
        break;
      case 'RETURN_REQUESTED':
        bg = Colors.orange[50]!;
        fg = Colors.orange[800]!;
        label = 'Return Requested';
        break;
      case 'RETURNED':
        bg = Colors.blue[50]!;
        fg = Colors.blue[800]!;
        label = 'Returned';
        break;
      case 'CREATION_FAILED':
        bg = Colors.red[50]!;
        fg = Colors.red[800]!;
        label = 'Failed';
        break;
      default:
        bg = Colors.grey[100]!;
        fg = Colors.grey[800]!;
        label = status;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        label,
        style: TextStyle(color: fg, fontWeight: FontWeight.bold, fontSize: 12),
      ),
    );
  }
}

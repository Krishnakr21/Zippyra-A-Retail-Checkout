import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/returnable_item.dart';

class OrderItemRow extends StatelessWidget {
  final ReturnableItem item;

  const OrderItemRow({
    super.key,
    required this.item,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.name,
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                const SizedBox(height: 2),
                Text(
                  '${item.qty} x ${CurrencyFormatter.formatPaise(item.pricePaise)}',
                  style: const TextStyle(color: Colors.grey, fontSize: 12),
                ),
                if (!item.isReturnable)
                  const Padding(
                    padding: EdgeInsets.only(top: 2.0),
                    child: Text(
                      'Non-returnable item',
                      style: TextStyle(color: Colors.red, fontSize: 11, fontWeight: FontWeight.bold),
                    ),
                  ),
              ],
            ),
          ),
          Text(
            CurrencyFormatter.formatPaise(item.pricePaise * item.qty),
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
          ),
        ],
      ),
    );
  }
}

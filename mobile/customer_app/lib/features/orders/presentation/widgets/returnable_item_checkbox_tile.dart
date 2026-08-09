import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/returnable_item.dart';

class ReturnableItemCheckboxTile extends StatelessWidget {
  final ReturnableItem item;
  final bool isSelected;
  final int selectedQty;
  final ValueChanged<bool?> onCheckboxChanged;
  final ValueChanged<int> onQtyChanged;

  const ReturnableItemCheckboxTile({
    super.key,
    required this.item,
    required this.isSelected,
    required this.selectedQty,
    required this.onCheckboxChanged,
    required this.onQtyChanged,
  });

  @override
  Widget build(BuildContext context) {
    final maxQty = item.remainingReturnableQty;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        child: Row(
          children: [
            Checkbox(
              value: isSelected,
              activeColor: ZippyraColors.primary,
              onChanged: maxQty > 0 ? onCheckboxChanged : null,
            ),
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
                    'Price: ${CurrencyFormatter.formatPaise(item.pricePaise)} • Purchased: ${item.qty}',
                    style: const TextStyle(color: Colors.grey, fontSize: 12),
                  ),
                ],
              ),
            ),
            if (isSelected && maxQty > 1)
              DropdownButton<int>(
                value: selectedQty.clamp(1, maxQty),
                items: List.generate(maxQty, (index) => index + 1)
                    .map((qty) => DropdownMenuItem<int>(
                          value: qty,
                          child: Text('$qty'),
                        ))
                    .toList(),
                onChanged: (val) {
                  if (val != null) onQtyChanged(val);
                },
              ),
          ],
        ),
      ),
    );
  }
}

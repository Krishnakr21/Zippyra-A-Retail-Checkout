import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/cart_summary.dart';

class CartTotalsSummary extends StatelessWidget {
  final CartSummary summary;

  const CartTotalsSummary({super.key, required this.summary});

  @override
  Widget build(BuildContext context) {
    final showIntraStateTax = summary.igstPaise <= 0 && (summary.cgstPaise > 0 || summary.sgstPaise > 0);
    final showInterStateTax = summary.igstPaise > 0;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      elevation: 1,
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Bill Details',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const Divider(height: 20),

            _buildRow('Subtotal', CurrencyFormatter.formatPaise(summary.subtotalPaise)),

            if (summary.discountPaise > 0)
              _buildRow(
                'Discount',
                '-${CurrencyFormatter.formatPaise(summary.discountPaise)}',
                valueColor: Colors.green,
              ),

            if (showIntraStateTax) ...[
              _buildRow('CGST', CurrencyFormatter.formatPaise(summary.cgstPaise)),
              _buildRow('SGST', CurrencyFormatter.formatPaise(summary.sgstPaise)),
            ],

            if (showInterStateTax)
              _buildRow('IGST', CurrencyFormatter.formatPaise(summary.igstPaise)),

            const Divider(height: 20),

            _buildRow(
              'Grand Total',
              CurrencyFormatter.formatPaise(summary.totalPaise),
              isBold: true,
              fontSize: 16,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRow(String label, String value, {bool isBold = false, double fontSize = 14, Color? valueColor}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: fontSize,
              fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
              color: isBold ? Colors.black : Colors.grey[700],
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: fontSize,
              fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
              color: valueColor ?? (isBold ? Colors.black : Colors.grey[800]),
            ),
          ),
        ],
      ),
    );
  }
}

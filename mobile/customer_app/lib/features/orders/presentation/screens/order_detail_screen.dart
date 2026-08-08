import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/order_detail_bloc.dart';
import '../widgets/order_item_row.dart';
import '../widgets/order_status_chip.dart';

class OrderDetailScreen extends StatefulWidget {
  final String orderId;

  const OrderDetailScreen({
    super.key,
    required this.orderId,
  });

  @override
  State<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends State<OrderDetailScreen> {
  @override
  void initState() {
    super.initState();
    context.read<OrderDetailBloc>().add(OrderDetailRequested(widget.orderId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Order Receipt'),
      ),
      body: BlocBuilder<OrderDetailBloc, OrderDetailState>(
        builder: (context, state) {
          if (state is OrderDetailLoading) {
            return const Center(child: CircularProgressIndicator(color: ZippyraColors.primary));
          }

          if (state is OrderDetailError) {
            return Center(
              child: Text('Error: ${state.message}'),
            );
          }

          if (state is OrderDetailLoaded) {
            final order = state.order;

            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                // Header Card
                Card(
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text('Order #${order.id.substring(0, order.id.length.clamp(0, 12))}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                            OrderStatusChip(status: order.status),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text('Payment Method: ${order.paymentMethod}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                        Text('Date: ${order.createdAt.day}/${order.createdAt.month}/${order.createdAt.year}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 16),

                // Itemized List Card
                Card(
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Items Purchased', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        const Divider(height: 24),
                        ...order.items.map((item) => OrderItemRow(item: item)),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 16),

                // GST Totals Breakdown Card
                Card(
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      children: [
                        _buildRow('Subtotal', CurrencyFormatter.formatPaise(order.subtotalPaise)),
                        if (order.discountPaise > 0)
                          _buildRow('Discount', '- ${CurrencyFormatter.formatPaise(order.discountPaise)}', color: Colors.green),
                        if (order.cgstPaise > 0)
                          _buildRow('CGST (9%)', CurrencyFormatter.formatPaise(order.cgstPaise)),
                        if (order.sgstPaise > 0)
                          _buildRow('SGST (9%)', CurrencyFormatter.formatPaise(order.sgstPaise)),
                        if (order.igstPaise > 0)
                          _buildRow('IGST (18%)', CurrencyFormatter.formatPaise(order.igstPaise)),
                        const Divider(height: 20),
                        _buildRow('Total Paid', CurrencyFormatter.formatPaise(order.totalPaise), isBold: true, fontSize: 16),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 16),

                // GST e-Invoice QR Code if available
                if (order.irnQrCodeBase64 != null && order.irnQrCodeBase64!.isNotEmpty)
                  Card(
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        children: [
                          const Text('GST e-Invoice Signed QR', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                          const SizedBox(height: 12),
                          SizedBox(
                            width: 150,
                            height: 150,
                            child: Image.memory(base64Decode(order.irnQrCodeBase64!.replaceFirst(RegExp(r'data:image/[^;]+;base64,'), ''))),
                          ),
                        ],
                      ),
                    ),
                  ),

                const SizedBox(height: 16),

                // Download Invoice Button
                if (order.signedInvoiceUrl != null && order.signedInvoiceUrl!.isNotEmpty)
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: OutlinedButton.icon(
                      style: OutlinedButton.styleFrom(
                        side: const BorderSide(color: ZippyraColors.primary),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      icon: const Icon(Icons.picture_as_pdf, color: ZippyraColors.primary),
                      label: const Text('Download Invoice PDF', style: TextStyle(color: ZippyraColors.primary, fontWeight: FontWeight.bold)),
                      onPressed: () {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('Opening Invoice URL: ${order.signedInvoiceUrl}')),
                        );
                      },
                    ),
                  ),

                const SizedBox(height: 12),

                // Request Return Button (visible only if within 24h & has returnable items)
                if (order.canRequestReturn)
                  SizedBox(
                    width: double.infinity,
                    height: 50,
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.orange[800],
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      icon: const Icon(Icons.assignment_return, color: Colors.white),
                      label: const Text('Request Return', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
                      onPressed: () {
                        context.push('/order/${order.id}/return', extra: order);
                      },
                    ),
                  ),
              ],
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildRow(String label, String value, {bool isBold = false, double fontSize = 14, Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(fontWeight: isBold ? FontWeight.bold : FontWeight.normal, fontSize: fontSize, color: color)),
          Text(value, style: TextStyle(fontWeight: isBold ? FontWeight.bold : FontWeight.normal, fontSize: fontSize, color: color)),
        ],
      ),
    );
  }
}

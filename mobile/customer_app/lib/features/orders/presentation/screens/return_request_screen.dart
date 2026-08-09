import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/order_detail.dart';
import '../bloc/order_detail_bloc.dart';
import '../widgets/returnable_item_checkbox_tile.dart';

class ReturnRequestScreen extends StatefulWidget {
  final OrderDetail order;

  const ReturnRequestScreen({
    super.key,
    required this.order,
  });

  @override
  State<ReturnRequestScreen> createState() => _ReturnRequestScreenState();
}

class _ReturnRequestScreenState extends State<ReturnRequestScreen> {
  final Set<String> _selectedBarcodes = {};
  final Map<String, int> _selectedQty = {};
  String _reason = 'DAMAGED';

  @override
  Widget build(BuildContext context) {
    final returnableItems = widget.order.items.where((i) => i.remainingReturnableQty > 0).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Request Item Return'),
      ),
      body: BlocConsumer<OrderDetailBloc, OrderDetailState>(
        listener: (context, state) {
          if (state is ReturnSubmitted) {
            context.go('/order/${widget.order.id}/return/confirmation');
          }
        },
        builder: (context, state) {
          return Column(
            children: [
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    if (state is ReturnFailed)
                      Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.red[50],
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: Colors.red[300]!),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.error_outline, color: Colors.red),
                            const SizedBox(width: 10),
                            Expanded(
                              child: Text(
                                state.message,
                                style: const TextStyle(color: Colors.red, fontWeight: FontWeight.bold, fontSize: 13),
                              ),
                            ),
                          ],
                        ),
                      ),

                    const Text(
                      'Select Items to Return',
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                    ),
                    const SizedBox(height: 12),

                    ...returnableItems.map((item) {
                      final isSelected = _selectedBarcodes.contains(item.barcode);
                      return ReturnableItemCheckboxTile(
                        item: item,
                        isSelected: isSelected,
                        selectedQty: _selectedQty[item.barcode] ?? 1,
                        onCheckboxChanged: (val) {
                          setState(() {
                            if (val == true) {
                              _selectedBarcodes.add(item.barcode);
                              _selectedQty[item.barcode] = 1;
                            } else {
                              _selectedBarcodes.remove(item.barcode);
                              _selectedQty.remove(item.barcode);
                            }
                          });
                        },
                        onQtyChanged: (qty) {
                          setState(() {
                            _selectedQty[item.barcode] = qty;
                          });
                        },
                      );
                    }),

                    const SizedBox(height: 16),

                    const Text(
                      'Reason for Return',
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                    ),
                    const SizedBox(height: 8),

                    DropdownButtonFormField<String>(
                      value: _reason,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                      ),
                      items: const [
                        DropdownMenuItem(value: 'DAMAGED', child: Text('Item Damaged / Defective')),
                        DropdownMenuItem(value: 'WRONG_ITEM', child: Text('Wrong Item Delivered')),
                        DropdownMenuItem(value: 'OTHER', child: Text('Other Reason')),
                      ],
                      onChanged: (val) {
                        if (val != null) setState(() => _reason = val);
                      },
                    ),
                  ],
                ),
              ),

              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  boxShadow: [
                    BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 10, offset: const Offset(0, -5)),
                  ],
                ),
                child: SafeArea(
                  child: SizedBox(
                    width: double.infinity,
                    height: 50,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: ZippyraColors.primary,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      onPressed: (_selectedBarcodes.isEmpty || state is ReturnSubmitting)
                          ? null
                          : () {
                              final List<String> barcodesToSubmit = [];
                              for (final barcode in _selectedBarcodes) {
                                final qty = _selectedQty[barcode] ?? 1;
                                for (int i = 0; i < qty; i++) {
                                  barcodesToSubmit.add(barcode);
                                }
                              }
                              context.read<OrderDetailBloc>().add(ReturnRequested(
                                    orderId: widget.order.id,
                                    itemBarcodes: barcodesToSubmit,
                                    reason: _reason,
                                  ));
                            },
                      child: state is ReturnSubmitting
                          ? const CircularProgressIndicator(color: Colors.white)
                          : const Text(
                              'Submit Return Request',
                              style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                            ),
                    ),
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

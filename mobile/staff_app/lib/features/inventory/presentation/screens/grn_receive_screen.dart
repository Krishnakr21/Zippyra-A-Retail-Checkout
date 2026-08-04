import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/grn_bloc.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class GrnReceiveScreen extends StatefulWidget {
  final String? poId;

  const GrnReceiveScreen({super.key, this.poId});

  @override
  State<GrnReceiveScreen> createState() => _GrnReceiveScreenState();
}

class _GrnReceiveScreenState extends State<GrnReceiveScreen> {
  final _invoiceRefController = TextEditingController();
  final _barcodeController = TextEditingController();
  final _qtyController = TextEditingController();
  final List<Map<String, dynamic>> _items = [];

  @override
  void dispose() {
    _invoiceRefController.dispose();
    _barcodeController.dispose();
    _qtyController.dispose();
    super.dispose();
  }

  void _addItem() {
    final barcode = _barcodeController.text.trim();
    final qty = int.tryParse(_qtyController.text.trim()) ?? 0;
    if (barcode.isNotEmpty && qty > 0) {
      setState(() {
        _items.add({
          'barcode': barcode,
          'qty_received': qty,
          'unit_cost_paise': 1000,
        });
      });
      _barcodeController.clear();
      _qtyController.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    String storeId = 'store-001';
    if (authState is AuthAuthenticated) {
      storeId = authState.storeId;
    }

    final isAdHoc = widget.poId == null || widget.poId == 'new';

    return Scaffold(
      appBar: AppBar(
        title: Text(isAdHoc ? 'Create Ad-hoc GRN' : 'Receive PO #${widget.poId!.substring(0, 8)}'),
      ),
      body: BlocListener<GrnBloc, GrnState>(
        listener: (context, state) {
          if (state is GrnCreated) {
            final grnId = state.grnData['id'] as String? ?? 'grn-id';
            final status = state.grnData['status'] as String? ?? 'DRAFT';
            if (status == 'QC_PENDING') {
              context.push('/home/inventory/qc/$grnId');
            } else {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('GRN Created Successfully')),
              );
              context.pop();
            }
          } else if (state is GrnError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
          }
        },
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              ZInput(
                label: 'Invoice Ref',
                hint: 'Vendor Invoice Reference (Optional)',
                controller: _invoiceRefController,
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    flex: 2,
                    child: ZInput(
                      label: 'Barcode',
                      hint: 'Barcode',
                      controller: _barcodeController,
                      keyboardType: TextInputType.number,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    flex: 1,
                    child: ZInput(
                      label: 'Qty',
                      hint: 'Qty',
                      controller: _qtyController,
                      keyboardType: TextInputType.number,
                    ),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(onPressed: _addItem, child: const Text('Add')),
                ],
              ),
              const SizedBox(height: 16),
              const Text('Items Received:', style: TextStyle(fontWeight: FontWeight.bold)),
              Expanded(
                child: _items.isEmpty
                    ? const Center(child: Text('No items added yet', style: TextStyle(color: Colors.grey)))
                    : ListView.builder(
                        itemCount: _items.length,
                        itemBuilder: (context, index) {
                          final item = _items[index];
                          return Card(
                            child: ListTile(
                              title: Text(item['barcode'] as String),
                              trailing: Text('Qty: ${item['qty_received']}'),
                            ),
                          );
                        },
                      ),
              ),
              BlocBuilder<GrnBloc, GrnState>(
                builder: (context, state) {
                  return ZButton(
                    label: 'Create GRN',
                    isLoading: state is GrnLoading,
                    onPressed: _items.isEmpty
                        ? () {}
                        : () {
                            context.read<GrnBloc>().add(GrnCreateRequested(
                                  storeId: storeId,
                                  poId: isAdHoc ? null : widget.poId,
                                  vendorInvoiceRef: _invoiceRefController.text.trim(),
                                  items: _items,
                                ));
                          },
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

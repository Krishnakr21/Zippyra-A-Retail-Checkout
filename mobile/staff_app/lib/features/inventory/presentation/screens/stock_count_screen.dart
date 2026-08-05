import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/stock_count_bloc.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class StockCountScreen extends StatefulWidget {
  const StockCountScreen({super.key});

  @override
  State<StockCountScreen> createState() => _StockCountScreenState();
}

class _StockCountScreenState extends State<StockCountScreen> {
  final _manualBarcodeController = TextEditingController();

  @override
  void dispose() {
    _manualBarcodeController.dispose();
    super.dispose();
  }

  void _addManualBarcode() {
    final code = _manualBarcodeController.text.trim();
    if (code.isNotEmpty) {
      context.read<StockCountBloc>().add(ItemScanned(barcode: code));
      _manualBarcodeController.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    String storeId = 'store-001';
    if (authState is AuthAuthenticated) {
      storeId = authState.storeId;
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Stock Count')),
      body: BlocListener<StockCountBloc, StockCountState>(
        listener: (context, state) {
          if (state is StockCountSubmittedWithVariance) {
            showDialog(
              context: context,
              builder: (ctx) => AlertDialog(
                title: const Text('Count Submitted'),
                content: Text(
                  'Submitted successfully.\nDiscrepancies found: ${state.discrepanciesCount}',
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.of(ctx).pop(),
                    child: const Text('OK'),
                  ),
                ],
              ),
            );
          } else if (state is StockCountQueuedOffline) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Saved offline — will sync when connected'),
                backgroundColor: Colors.orange,
              ),
            );
          }
        },
        child: Column(
          children: [
            // Manual Barcode Entry Card
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Row(
                children: [
                  Expanded(
                    child: ZInput(
                      label: 'Manual Barcode',
                      hint: 'Enter barcode manually...',
                      controller: _manualBarcodeController,
                      keyboardType: TextInputType.number,
                    ),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: _addManualBarcode,
                    child: const Text('Add'),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),

            // Itemized Counted List
            Expanded(
              child: BlocBuilder<StockCountBloc, StockCountState>(
                builder: (context, state) {
                  final entries = (state is StockCountLoaded)
                      ? state.entries
                      : (state is StockCountSubmitting ? state.entries : []);

                  if (entries.isEmpty) {
                    return const Center(
                      child: Text(
                        'Scan or enter barcodes to begin count',
                        style: TextStyle(color: Colors.grey),
                      ),
                    );
                  }

                  return ListView.builder(
                    padding: const EdgeInsets.all(16.0),
                    itemCount: entries.length,
                    itemBuilder: (context, index) {
                      final item = entries[index];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          title: Text(item.name, style: const TextStyle(fontWeight: FontWeight.bold)),
                          subtitle: Text('Barcode: ${item.barcode}'),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              IconButton(
                                icon: const Icon(Icons.remove_circle_outline),
                                onPressed: () {
                                  context.read<StockCountBloc>().add(
                                        ItemCountEdited(barcode: item.barcode, qty: item.countedQty - 1),
                                      );
                                },
                              ),
                              Text(
                                '${item.countedQty}',
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                              ),
                              IconButton(
                                icon: const Icon(Icons.add_circle_outline),
                                onPressed: () {
                                  context.read<StockCountBloc>().add(
                                        ItemCountEdited(barcode: item.barcode, qty: item.countedQty + 1),
                                      );
                                },
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  );
                },
              ),
            ),

            // Submit Button Footer
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: BlocBuilder<StockCountBloc, StockCountState>(
                builder: (context, state) {
                  final isSubmitting = state is StockCountSubmitting;
                  return ZButton(
                    label: 'Submit Count',
                    isLoading: isSubmitting,
                    onPressed: () {
                      context.read<StockCountBloc>().add(CountSubmitted(storeId));
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

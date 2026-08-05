import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/grn_bloc.dart';

class QcReviewScreen extends StatefulWidget {
  final String grnId;

  const QcReviewScreen({super.key, required this.grnId});

  @override
  State<QcReviewScreen> createState() => _QcReviewScreenState();
}

class _QcReviewScreenState extends State<QcReviewScreen> {
  final Map<String, String> _qcStatuses = {};

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('GRN Quality Control Review')),
      body: BlocListener<GrnBloc, GrnState>(
        listener: (context, state) {
          if (state is GrnCompleted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('GRN Completed & Stock Applied Successfully')),
            );
            context.go('/home/inventory');
          } else if (state is QcIncomplete) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('QC Incomplete: Please make QC decisions for all line items'),
                backgroundColor: Colors.red,
              ),
            );
          } else if (state is GrnAlreadyCompleted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('GRN Already Completed')),
            );
            context.go('/home/inventory');
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
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('GRN ID: ${widget.grnId}', style: const TextStyle(fontWeight: FontWeight.bold)),
                      const SizedBox(height: 4),
                      const Text('Review item quality before applying stock to inventory', style: TextStyle(color: Colors.grey, fontSize: 12)),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),

              Expanded(
                child: ListView(
                  children: [
                    _buildQcItemTile('item-1', 'Sample Item 1', 10),
                    _buildQcItemTile('item-2', 'Sample Item 2', 5),
                  ],
                ),
              ),

              BlocBuilder<GrnBloc, GrnState>(
                builder: (context, state) {
                  return ZButton(
                    label: 'Complete GRN',
                    isLoading: state is GrnLoading,
                    onPressed: () {
                      final updates = _qcStatuses.entries
                          .map((e) => {
                                'grn_line_item_id': e.key,
                                'qc_status': e.value,
                              })
                          .toList();

                      context.read<GrnBloc>().add(QcDecisionSubmitted(
                            grnId: widget.grnId,
                            lineItemUpdates: updates,
                          ));

                      context.read<GrnBloc>().add(GrnCompleteRequested(widget.grnId));
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

  Widget _buildQcItemTile(String itemId, String title, int qty) {
    final status = _qcStatuses[itemId] ?? 'PASSED';
    _qcStatuses[itemId] = status;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                Text('Qty: $qty', style: const TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: ChoiceChip(
                    label: const Text('PASSED'),
                    selected: status == 'PASSED',
                    selectedColor: Colors.green[100],
                    onSelected: (val) {
                      if (val) {
                        setState(() => _qcStatuses[itemId] = 'PASSED');
                      }
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: ChoiceChip(
                    label: const Text('REJECTED'),
                    selected: status == 'REJECTED',
                    selectedColor: Colors.red[100],
                    onSelected: (val) {
                      if (val) {
                        setState(() => _qcStatuses[itemId] = 'REJECTED');
                      }
                    },
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/low_stock_bloc.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class LowStockScreen extends StatefulWidget {
  const LowStockScreen({super.key});

  @override
  State<LowStockScreen> createState() => _LowStockScreenState();
}

class _LowStockScreenState extends State<LowStockScreen> {
  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    final authState = context.read<AuthBloc>().state;
    String storeId = 'store-001';
    if (authState is AuthAuthenticated) {
      storeId = authState.storeId;
    }
    context.read<LowStockBloc>().add(LowStockRequested(storeId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Low Stock Alerts'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _load,
          ),
        ],
      ),
      body: BlocBuilder<LowStockBloc, LowStockState>(
        builder: (context, state) {
          if (state is LowStockLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is LowStockError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 12),
                  Text(state.message),
                  const SizedBox(height: 16),
                  ElevatedButton(onPressed: _load, child: const Text('Retry')),
                ],
              ),
            );
          }

          if (state is LowStockLoaded) {
            if (state.items.isEmpty) {
              return const Center(
                child: Padding(
                  padding: EdgeInsets.all(24.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.check_circle_outline, size: 64, color: Colors.green),
                      SizedBox(height: 16),
                      Text(
                        'No low-stock items right now.',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                ),
              );
            }

            return RefreshIndicator(
              onRefresh: () async => _load(),
              child: ListView.builder(
                padding: const EdgeInsets.all(16.0),
                itemCount: state.items.length,
                itemBuilder: (context, index) {
                  final item = state.items[index];
                  return Card(
                    margin: const EdgeInsets.only(bottom: 12),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: ListTile(
                      leading: const CircleAvatar(
                        backgroundColor: Colors.amber,
                        child: Icon(Icons.warning_amber, color: Colors.white),
                      ),
                      title: Text(
                        item.productName,
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                      subtitle: Text(
                        'Barcode: ${item.barcode} • Threshold: ${item.reorderPoint}',
                        style: const TextStyle(fontSize: 12, color: Colors.grey),
                      ),
                      trailing: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(
                            '${item.onHandQty} left',
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Colors.red),
                          ),
                          Text(
                            'Reorder ${item.reorderQty}',
                            style: const TextStyle(fontSize: 11, color: Colors.grey),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }
}

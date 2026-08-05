import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/grn_bloc.dart';
import '../../../auth/presentation/bloc/auth_bloc.dart';

class GrnListScreen extends StatefulWidget {
  const GrnListScreen({super.key});

  @override
  State<GrnListScreen> createState() => _GrnListScreenState();
}

class _GrnListScreenState extends State<GrnListScreen> {
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
    context.read<GrnBloc>().add(GrnListRequested(storeId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('GRN Receive'),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: _load),
        ],
      ),
      body: Column(
        children: [
          // Action Banner for Ad-hoc GRN
          Container(
            padding: const EdgeInsets.all(16.0),
            color: Colors.blue[50],
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Unplanned Delivery?', style: TextStyle(fontWeight: FontWeight.bold)),
                    Text('Receive goods without a Purchase Order', style: TextStyle(fontSize: 12, color: Colors.grey)),
                  ],
                ),
                ElevatedButton(
                  onPressed: () {
                    context.push('/home/inventory/grn/new');
                  },
                  child: const Text('Ad-hoc GRN'),
                ),
              ],
            ),
          ),
          const Divider(height: 1),

          Expanded(
            child: BlocBuilder<GrnBloc, GrnState>(
              builder: (context, state) {
                if (state is GrnLoading) {
                  return const Center(child: CircularProgressIndicator());
                }

                if (state is GrnError) {
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

                if (state is GrnListLoaded) {
                  if (state.pos.isEmpty) {
                    return const Center(
                      child: Text('No pending submitted Purchase Orders to receive'),
                    );
                  }

                  return ListView.builder(
                    padding: const EdgeInsets.all(16.0),
                    itemCount: state.pos.length,
                    itemBuilder: (context, index) {
                      final po = state.pos[index];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 12),
                        child: ListTile(
                          leading: const CircleAvatar(child: Icon(Icons.receipt_long)),
                          title: Text(po.vendorName, style: const TextStyle(fontWeight: FontWeight.bold)),
                          subtitle: Text('PO #${po.id.substring(0, 8)} • ${po.status}'),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () {
                            context.push('/home/inventory/grn/${po.id}');
                          },
                        ),
                      );
                    },
                  );
                }

                return const SizedBox.shrink();
              },
            ),
          ),
        ],
      ),
    );
  }
}

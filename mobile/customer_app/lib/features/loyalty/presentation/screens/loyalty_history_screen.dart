import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../cubit/loyalty_history_cubit.dart';
import '../widgets/ledger_entry_tile.dart';

class LoyaltyHistoryScreen extends StatefulWidget {
  const LoyaltyHistoryScreen({super.key});

  @override
  State<LoyaltyHistoryScreen> createState() => _LoyaltyHistoryScreenState();
}

class _LoyaltyHistoryScreenState extends State<LoyaltyHistoryScreen> {
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    context.read<LoyaltyHistoryCubit>().fetchHistory();
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      context.read<LoyaltyHistoryCubit>().fetchNextPage();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Points History'),
      ),
      body: BlocBuilder<LoyaltyHistoryCubit, LoyaltyHistoryState>(
        builder: (context, state) {
          if (state is LoyaltyHistoryLoading && state is! LoyaltyHistoryLoaded) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is LoyaltyHistoryError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 12),
                  Text(state.message),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () {
                      context.read<LoyaltyHistoryCubit>().fetchHistory();
                    },
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (state is LoyaltyHistoryLoaded) {
            if (state.items.isEmpty) {
              return const Center(
                child: Padding(
                  padding: EdgeInsets.all(24.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.history_toggle_off, size: 64, color: Colors.grey),
                      SizedBox(height: 16),
                      Text(
                        'No points transactions yet',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      SizedBox(height: 8),
                      Text(
                        'Earn points on your first order with us!',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey, fontSize: 14),
                      ),
                    ],
                  ),
                ),
              );
            }

            return ListView.builder(
              controller: _scrollController,
              itemCount: state.items.length + (state.hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index == state.items.length) {
                  return const Padding(
                    padding: EdgeInsets.all(16.0),
                    child: Center(child: CircularProgressIndicator()),
                  );
                }

                final entry = state.items[index];
                return LedgerEntryTile(entry: entry);
              },
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }
}

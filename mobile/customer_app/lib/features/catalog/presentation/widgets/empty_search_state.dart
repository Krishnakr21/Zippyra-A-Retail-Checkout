import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';

class EmptySearchState extends StatelessWidget {
  final String query;

  const EmptySearchState({super.key, required this.query});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.search_off, size: 72, color: Colors.grey),
            const SizedBox(height: 16),
            Text(
              query.isNotEmpty ? 'No products found for "$query"' : 'No products found',
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            const Text(
              'Try searching with a different term or browse categories.',
              textAlign: TextAlign.center,
              style: TextStyle(color: ZippyraColors.textSecondary),
            ),
          ],
        ),
      ),
    );
  }
}

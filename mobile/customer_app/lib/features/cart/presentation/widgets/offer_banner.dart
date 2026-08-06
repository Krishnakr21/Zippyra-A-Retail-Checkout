import 'package:flutter/material.dart';

class OfferBanner extends StatelessWidget {
  final List<String> offers;

  const OfferBanner({super.key, required this.offers});

  @override
  Widget build(BuildContext context) {
    if (offers.isEmpty) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.green[50],
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.green[300]!),
      ),
      child: Row(
        children: [
          const Icon(Icons.local_offer, color: Colors.green, size: 20),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              offers.join(' • '),
              style: const TextStyle(
                color: Colors.green,
                fontWeight: FontWeight.bold,
                fontSize: 13,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

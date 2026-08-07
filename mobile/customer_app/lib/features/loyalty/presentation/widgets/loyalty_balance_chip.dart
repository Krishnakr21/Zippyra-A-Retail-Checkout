import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import '../bloc/loyalty_bloc.dart';

class LoyaltyBalanceChip extends StatelessWidget {
  const LoyaltyBalanceChip({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<LoyaltyBloc, LoyaltyState>(
      builder: (context, state) {
        int balance = 0;
        if (state is LoyaltyLoaded) {
          balance = state.pointsBalance;
        }

        return InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: () {
            context.push('/loyalty');
          },
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: Colors.amber[50],
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.amber[300]!, width: 1),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.star, color: Colors.amber, size: 16),
                const SizedBox(width: 4),
                Text(
                  '$balance pts',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 12,
                    color: Colors.amber[900],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

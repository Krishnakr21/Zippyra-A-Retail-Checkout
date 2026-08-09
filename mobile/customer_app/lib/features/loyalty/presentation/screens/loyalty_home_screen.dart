import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/loyalty_bloc.dart';
import '../widgets/tier_badge.dart';
import '../widgets/tier_progress_bar.dart';

class LoyaltyHomeScreen extends StatefulWidget {
  const LoyaltyHomeScreen({super.key});

  @override
  State<LoyaltyHomeScreen> createState() => _LoyaltyHomeScreenState();
}

class _LoyaltyHomeScreenState extends State<LoyaltyHomeScreen> {
  @override
  void initState() {
    super.initState();
    context.read<LoyaltyBloc>().add(const LoyaltyBalanceRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Rewards & Loyalty'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              context.read<LoyaltyBloc>().add(const LoyaltyBalanceRequested(refresh: true));
            },
          ),
        ],
      ),
      body: BlocBuilder<LoyaltyBloc, LoyaltyState>(
        builder: (context, state) {
          if (state is LoyaltyLoading && state is! LoyaltyLoaded) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is LoyaltyError) {
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
                      context.read<LoyaltyBloc>().add(const LoyaltyBalanceRequested(refresh: true));
                    },
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          int pointsBalance = 0;
          String tier = 'BRONZE';
          String tierDisplayName = 'Bronze Tier';
          int lifetimePoints = 0;
          int? pointsToNextTier = 5000;
          String? nextTierName = 'Silver Tier';

          if (state is LoyaltyLoaded) {
            pointsBalance = state.pointsBalance;
            tier = state.tier;
            tierDisplayName = state.tierDisplayName;
            lifetimePoints = state.lifetimePointsEarned;
            pointsToNextTier = state.pointsToNextTier;
            nextTierName = state.nextTierName;
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(20.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header Card
                Card(
                  elevation: 3,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            TierBadge(tier: tier, displayName: tierDisplayName),
                            const Icon(Icons.star, color: Colors.amber, size: 32),
                          ],
                        ),
                        const SizedBox(height: 20),
                        const Text(
                          'Available Points',
                          style: TextStyle(color: Colors.grey, fontSize: 13, fontWeight: FontWeight.w500),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '$pointsBalance',
                          style: const TextStyle(
                            fontSize: 36,
                            fontWeight: FontWeight.bold,
                            color: ZippyraColors.primary,
                          ),
                        ),
                        const SizedBox(height: 20),
                        TierProgressBar(
                          lifetimePointsEarned: lifetimePoints,
                          pointsToNextTier: pointsToNextTier,
                          nextTierName: nextTierName,
                        ),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 24),

                // Navigation Shortcuts
                Card(
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  child: Column(
                    children: [
                      ListTile(
                        leading: const Icon(Icons.history, color: ZippyraColors.primary),
                        title: const Text('Transaction History', style: TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: const Text('View points earned, redeemed, and reversed'),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () {
                          context.push('/loyalty/history');
                        },
                      ),
                      const Divider(height: 1),
                      ListTile(
                        leading: const Icon(Icons.military_tech_outlined, color: Colors.amber),
                        title: const Text('Tier Benefits & Ladder', style: TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: const Text('Learn how multipliers and tiers work'),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () {
                          context.push('/loyalty/tiers');
                        },
                      ),
                      const Divider(height: 1),
                      ListTile(
                        leading: const Icon(Icons.card_giftcard, color: Color(0xFF6366F1)),
                        title: const Text('Refer & Earn', style: TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: const Text('Invite friends & earn bonus points'),
                        trailing: const Icon(Icons.chevron_right),
                        onTap: () {
                          context.push('/loyalty/referral');
                        },
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

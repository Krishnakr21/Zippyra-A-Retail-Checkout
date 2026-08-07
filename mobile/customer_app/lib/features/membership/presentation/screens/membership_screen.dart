import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../cubit/membership_cubit.dart';
import '../../domain/entities/subscription_plan.dart';

class MembershipScreen extends StatefulWidget {
  const MembershipScreen({super.key});

  @override
  State<MembershipScreen> createState() => _MembershipScreenState();
}

class _MembershipScreenState extends State<MembershipScreen> {
  @override
  void initState() {
    super.initState();
    context.read<MembershipCubit>().loadMembershipData();
  }

  void _showCancelDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: const Text('Cancel Subscription?'),
        content: const Text(
          'Your member perks (+0.5x loyalty bonus, free delivery) will remain active until the end of your current billing period. Future auto-renewals will be stopped.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogCtx),
            child: const Text('Keep Membership'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent),
            onPressed: () {
              Navigator.pop(dialogCtx);
              context.read<MembershipCubit>().cancelSubscription();
            },
            child: const Text('Confirm Cancel', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Smart Saver Membership'),
        elevation: 0,
      ),
      body: BlocBuilder<MembershipCubit, MembershipState>(
        builder: (context, state) {
          if (state is MembershipLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is MembershipError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(state.message, style: const TextStyle(color: Colors.redAccent)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => context.read<MembershipCubit>().loadMembershipData(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (state is MembershipLoaded) {
            final activeSub = state.activeSubscription;
            final plans = state.plans;

            return SingleChildScrollView(
              padding: const EdgeInsets.all(20.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Hero Header Banner
                  Container(
                    padding: const EdgeInsets.all(24.0),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF0F172A), Color(0xFF1E293B)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.2),
                          blurRadius: 15,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: const [
                            Icon(Icons.workspace_premium, color: Colors.amber, size: 40),
                            SizedBox(width: 8),
                            Text(
                              'SMART SAVER',
                              style: TextStyle(
                                color: Colors.amber,
                                fontSize: 24,
                                fontWeight: FontWeight.w800,
                                letterSpacing: 2,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        Text(
                          activeSub != null && activeSub.isActive
                              ? 'You are an active Smart Saver Member!'
                              : 'Unlock Premium Perks & Extra Loyalty Points',
                          style: const TextStyle(color: Colors.white70, fontSize: 14),
                          textAlign: TextAlign.center,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Active Subscription Card
                  if (activeSub != null && activeSub.isActive) ...[
                    Card(
                      elevation: 2,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      child: Padding(
                        padding: const EdgeInsets.all(20.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  activeSub.plan?.name ?? 'Smart Saver Plan',
                                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: Colors.green.shade100,
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: const Text(
                                    'ACTIVE',
                                    style: TextStyle(
                                      color: Colors.green,
                                      fontWeight: FontWeight.bold,
                                      fontSize: 12,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 12),
                            if (activeSub.currentPeriodEnd != null)
                              Text(
                                'Renews/Expires: ${activeSub.currentPeriodEnd!.day}/${activeSub.currentPeriodEnd!.month}/${activeSub.currentPeriodEnd!.year}',
                                style: const TextStyle(color: Colors.grey, fontSize: 13),
                              ),
                            const SizedBox(height: 20),
                            OutlinedButton.icon(
                              onPressed: () => _showCancelDialog(context),
                              icon: const Icon(Icons.cancel_outlined, color: Colors.redAccent),
                              label: const Text('Cancel Subscription', style: TextStyle(color: Colors.redAccent)),
                              style: OutlinedButton.styleFrom(
                                side: const BorderSide(color: Colors.redAccent),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ] else ...[
                    // Plan Selection Cards
                    const Text(
                      'CHOOSE YOUR PLAN',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 1.2,
                        color: Colors.grey,
                      ),
                    ),
                    const SizedBox(height: 12),
                    ...plans.map((plan) => _buildPlanCard(context, plan)),
                    const SizedBox(height: 24),
                  ],

                  // Benefits Breakdown
                  const Text(
                    'MEMBER BENEFITS',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 1.2,
                      color: Colors.grey,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _buildBenefitRow(
                    icon: Icons.bolt,
                    iconColor: Colors.amber,
                    title: '+0.5x Extra Loyalty Multiplier',
                    description: 'Earn 1.5x points on Bronze tier, 1.75x on Silver, 2.0x on Gold!',
                  ),
                  const SizedBox(height: 12),
                  _buildBenefitRow(
                    icon: Icons.local_shipping,
                    iconColor: Color(0xFF6366F1),
                    title: 'Free Delivery Equivalent Perks',
                    description: 'Zero convenience fees on express self-checkouts and orders.',
                  ),
                  const SizedBox(height: 12),
                  _buildBenefitRow(
                    icon: Icons.sell,
                    iconColor: Colors.green,
                    title: 'Exclusive Member Pricing',
                    description: 'Access special Smart Saver discounts across partner stores.',
                  ),
                ],
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildPlanCard(BuildContext context, SubscriptionPlan plan) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    plan.name,
                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    plan.formattedPrice,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: ZippyraColors.primary,
                    ),
                  ),
                ],
              ),
            ),
            ElevatedButton(
              onPressed: () {
                context.read<MembershipCubit>().subscribe(plan.id);
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: ZippyraColors.primary,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: const Text('Subscribe', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBenefitRow({
    required IconData icon,
    required Color iconColor,
    required String title,
    required String description,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CircleAvatar(
          radius: 18,
          backgroundColor: iconColor.withOpacity(0.15),
          child: Icon(icon, color: iconColor, size: 20),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
              ),
              const SizedBox(height: 2),
              Text(
                description,
                style: const TextStyle(fontSize: 13, color: Colors.grey),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

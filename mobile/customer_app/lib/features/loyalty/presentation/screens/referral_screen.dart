import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:share_plus/share_plus.dart';
import '../cubit/referral_cubit.dart';

class ReferralScreen extends StatefulWidget {
  const ReferralScreen({super.key});

  @override
  State<ReferralScreen> createState() => _ReferralScreenState();
}

class _ReferralScreenState extends State<ReferralScreen> {
  @override
  void initState() {
    super.initState();
    context.read<ReferralCubit>().loadReferralInfo();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Refer & Earn'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: BlocBuilder<ReferralCubit, ReferralState>(
        builder: (context, state) {
          if (state is ReferralLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is ReferralError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(state.message, style: const TextStyle(color: Colors.redAccent)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => context.read<ReferralCubit>().loadReferralInfo(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (state is ReferralLoaded) {
            final info = state.info;
            return SingleChildScrollView(
              padding: const EdgeInsets.all(20.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Hero Header Card
                  Container(
                    padding: const EdgeInsets.all(24.0),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF6366F1), Color(0xFF8B5CF6)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(20),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF6366F1).withOpacity(0.3),
                          blurRadius: 15,
                          offset: const Offset(0, 8),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        const Icon(Icons.card_giftcard, size: 56, color: Colors.white),
                        const SizedBox(height: 16),
                        const Text(
                          'Invite Friends, Earn Rewards!',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 22,
                            fontWeight: FontWeight.bold,
                          ),
                          textAlign: TextAlign.center,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Earn ${info.referrerRewardPoints} points for every friend who signs up and completes their first checkout. They get ${info.referredRewardPoints} bonus points too!',
                          style: const TextStyle(color: Colors.white70, fontSize: 14),
                          textAlign: TextAlign.center,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 28),

                  // Referral Code Card
                  const Text(
                    'YOUR REFERRAL CODE',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 1.2,
                      color: Colors.grey,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                    decoration: BoxDecoration(
                      color: Theme.of(context).cardColor,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: Colors.grey.withOpacity(0.2)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        SelectableText(
                          info.referralCode,
                          style: const TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                            letterSpacing: 3,
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.copy, color: Color(0xFF6366F1)),
                          onPressed: () {
                            Clipboard.setData(ClipboardData(text: info.referralCode));
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Referral code copied to clipboard!')),
                            );
                          },
                          tooltip: 'Copy Code',
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Share Button
                  ElevatedButton.icon(
                    onPressed: () {
                      Share.share(info.shareText);
                    },
                    icon: const Icon(Icons.share, color: Colors.white),
                    label: const Text(
                      'Share Code via WhatsApp / Apps',
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF25D366), // WhatsApp / Green
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                    ),
                  ),
                  const SizedBox(height: 32),

                  // How it works explainer
                  const Text(
                    'HOW IT WORKS',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 1.2,
                      color: Colors.grey,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _buildStepRow(
                    context,
                    step: '1',
                    title: 'Share your code',
                    subtitle: 'Send your code to friends, family, or colleagues.',
                  ),
                  const SizedBox(height: 12),
                  _buildStepRow(
                    context,
                    step: '2',
                    title: 'Friend signs up',
                    subtitle: 'Your friend enters your code during signup or onboarding.',
                  ),
                  const SizedBox(height: 12),
                  _buildStepRow(
                    context,
                    step: '3',
                    title: 'Get Rewarded',
                    subtitle: 'When they complete their first order, both of you receive bonus points!',
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

  Widget _buildStepRow(
    BuildContext context, {
    required String step,
    required String title,
    required String subtitle,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CircleAvatar(
          radius: 14,
          backgroundColor: const Color(0xFF6366F1).withOpacity(0.2),
          child: Text(
            step,
            style: const TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: Color(0xFF6366F1),
            ),
          ),
        ),
        const SizedBox(width: 12),
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
                subtitle,
                style: const TextStyle(fontSize: 13, color: Colors.grey),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

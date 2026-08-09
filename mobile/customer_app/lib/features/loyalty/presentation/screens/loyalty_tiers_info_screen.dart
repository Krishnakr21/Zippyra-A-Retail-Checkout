import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/tier_info.dart';
import '../../domain/usecases/get_tiers_info_use_case.dart';
import '../bloc/loyalty_bloc.dart';
import '../widgets/tier_badge.dart';

class LoyaltyTiersInfoScreen extends StatefulWidget {
  const LoyaltyTiersInfoScreen({super.key});

  @override
  State<LoyaltyTiersInfoScreen> createState() => _LoyaltyTiersInfoScreenState();
}

class _LoyaltyTiersInfoScreenState extends State<LoyaltyTiersInfoScreen> {
  List<TierInfo>? _tiers;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadTiers();
  }

  Future<void> _loadTiers() async {
    final useCase = context.read<GetTiersInfoUseCase>();
    try {
      final tiers = await useCase();
      if (mounted) {
        setState(() {
          _tiers = tiers;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    String currentTier = 'BRONZE';
    final loyaltyState = context.watch<LoyaltyBloc>().state;
    if (loyaltyState is LoyaltyLoaded) {
      currentTier = loyaltyState.tier;
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tier Benefits'),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Loyalty Tier Ladder',
                        style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                      ),
                      const SizedBox(height: 8),
                      const Text(
                        'Earn 1 base point for every ₹10 spent. Higher tiers earn bonus point multipliers on every checkout!',
                        style: TextStyle(color: Colors.grey, fontSize: 14),
                      ),
                      const SizedBox(height: 24),
                      ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        itemCount: _tiers?.length ?? 0,
                        itemBuilder: (context, index) {
                          final t = _tiers![index];
                          final isCurrent = t.tier.toUpperCase() == currentTier.toUpperCase();

                          return Container(
                            margin: const EdgeInsets.only(bottom: 12),
                            decoration: BoxDecoration(
                              color: isCurrent ? ZippyraColors.primary.withOpacity(0.06) : Colors.white,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: isCurrent ? ZippyraColors.primary : Colors.grey[300]!,
                                width: isCurrent ? 2 : 1,
                              ),
                            ),
                            child: ListTile(
                              leading: TierBadge(tier: t.tier, displayName: t.displayName),
                              title: Text(
                                '${t.earnMultiplier}x Points Multiplier',
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                              ),
                              subtitle: Text(
                                t.minLifetimePoints == 0
                                    ? 'Starting Tier'
                                    : 'Requires ${t.minLifetimePoints} lifetime points',
                                style: const TextStyle(fontSize: 12, color: Colors.grey),
                              ),
                              trailing: isCurrent
                                  ? Container(
                                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                      decoration: BoxDecoration(
                                        color: ZippyraColors.primary,
                                        borderRadius: BorderRadius.circular(6),
                                      ),
                                      child: const Text(
                                        'YOUR TIER',
                                        style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 10),
                                      ),
                                    )
                                  : null,
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ),
    );
  }
}

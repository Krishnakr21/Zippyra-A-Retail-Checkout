import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';

class TierProgressBar extends StatelessWidget {
  final int lifetimePointsEarned;
  final int? pointsToNextTier;
  final String? nextTierName;

  const TierProgressBar({
    super.key,
    required this.lifetimePointsEarned,
    this.pointsToNextTier,
    this.nextTierName,
  });

  @override
  Widget build(BuildContext context) {
    if (nextTierName == null || pointsToNextTier == null) {
      return const SizedBox.shrink(); // Top tier reached, hide progress bar
    }

    final totalTarget = lifetimePointsEarned + pointsToNextTier!;
    final progress = totalTarget > 0 ? (lifetimePointsEarned / totalTarget).clamp(0.0, 1.0) : 0.0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '${pointsToNextTier!} pts to $nextTierName',
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.black87),
            ),
            Text(
              '${(progress * 100).toInt()}%',
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13, color: Colors.grey),
            ),
          ],
        ),
        const SizedBox(height: 8),
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: LinearProgressIndicator(
            value: progress,
            minHeight: 10,
            backgroundColor: Colors.grey[200],
            valueColor: const AlwaysStoppedAnimation<Color>(ZippyraColors.primary),
          ),
        ),
      ],
    );
  }
}

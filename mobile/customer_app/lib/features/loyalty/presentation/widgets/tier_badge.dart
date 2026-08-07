import 'package:flutter/material.dart';

class TierBadge extends StatelessWidget {
  final String tier;
  final String displayName;
  final double fontSize;

  const TierBadge({
    super.key,
    required this.tier,
    required this.displayName,
    this.fontSize = 14.0,
  });

  @override
  Widget build(BuildContext context) {
    final t = tier.toUpperCase();
    Color bgColor;
    Color textColor;
    IconData icon;

    switch (t) {
      case 'PLATINUM':
        bgColor = Colors.deepPurple[100]!;
        textColor = Colors.deepPurple[900]!;
        icon = Icons.workspace_premium;
        break;
      case 'GOLD':
        bgColor = Colors.amber[100]!;
        textColor = Colors.amber[900]!;
        icon = Icons.military_tech;
        break;
      case 'SILVER':
        bgColor = Colors.grey[200]!;
        textColor = Colors.grey[800]!;
        icon = Icons.stars;
        break;
      case 'BRONZE':
      default:
        bgColor = Colors.brown[100]!;
        textColor = Colors.brown[900]!;
        icon = Icons.shield_outlined;
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: textColor, size: fontSize + 2),
          const SizedBox(width: 6),
          Text(
            displayName,
            style: TextStyle(
              color: textColor,
              fontWeight: FontWeight.bold,
              fontSize: fontSize,
            ),
          ),
        ],
      ),
    );
  }
}

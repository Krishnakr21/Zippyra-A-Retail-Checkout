import 'package:flutter/material.dart';
import '../theme/zippyra_theme.dart';

class ZBadge extends StatelessWidget {
  final String text;
  final Color color;

  const ZBadge({
    super.key,
    required this.text,
    this.color = ZippyraColors.primaryBlue,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

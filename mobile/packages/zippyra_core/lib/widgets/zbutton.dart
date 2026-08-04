import 'package:flutter/material.dart';
import '../theme/zippyra_theme.dart';

enum ZButtonType { primary, ghost, orange, green, red }

class ZButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;
  final ZButtonType type;
  final bool isLoading;
  final IconData? icon;

  const ZButton({
    super.key,
    required this.label,
    this.onPressed,
    this.type = ZButtonType.primary,
    this.isLoading = false,
    this.icon,
  });

  Color _getBgColor() {
    switch (type) {
      case ZButtonType.primary:
        return ZippyraColors.primaryBlue;
      case ZButtonType.orange:
        return ZippyraColors.accentOrange;
      case ZButtonType.green:
        return ZippyraColors.successGreen;
      case ZButtonType.red:
        return ZippyraColors.errorRed;
      case ZButtonType.ghost:
        return Colors.transparent;
    }
  }

  Color _getTextColor() {
    return type == ZButtonType.ghost ? ZippyraColors.primaryBlue : Colors.white;
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 52,
      width: double.infinity,
      child: ElevatedButton(
        style: ElevatedButton.styleFrom(
          backgroundColor: _getBgColor(),
          foregroundColor: _getTextColor(),
          elevation: type == ZButtonType.ghost ? 0 : 2,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: type == ZButtonType.ghost
                ? const BorderSide(color: ZippyraColors.primaryBlue, width: 1.5)
                : BorderSide.none,
          ),
        ),
        onPressed: isLoading ? null : onPressed,
        child: isLoading
            ? const SizedBox(
                height: 22,
                width: 22,
                child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5),
              )
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  if (icon != null) ...[Icon(icon, size: 20), const SizedBox(width: 8)],
                  Text(
                    label,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      color: _getTextColor(),
                    ),
                  ),
                ],
              ),
      ),
    );
  }
}

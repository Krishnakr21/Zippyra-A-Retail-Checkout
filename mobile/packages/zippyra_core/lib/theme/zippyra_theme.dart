import 'package:flutter/material.dart';

class ZippyraColors {
  static const Color primaryBlue = Color(0xFF1A5EAB);
  static const Color primary = primaryBlue;
  static const Color accentOrange = Color(0xFFF7941D);
  static const Color successGreen = Color(0xFF00C47A);
  static const Color errorRed = Color(0xFFFF3B3B);
  static const Color background = Color(0xFFF4F5FA);
  static const Color cardBg = Colors.white;
  static const Color textPrimary = Color(0xFF1E293B);
  static const Color textSecondary = Color(0xFF64748B);
  static const Color border = Color(0xFFE2E8F0);
}

class ZippyraTheme {
  static ThemeData get lightTheme {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: ZippyraColors.background,
      primaryColor: ZippyraColors.primaryBlue,
      colorScheme: ColorScheme.light(
        primary: ZippyraColors.primaryBlue,
        secondary: ZippyraColors.accentOrange,
        surface: ZippyraColors.cardBg,
        error: ZippyraColors.errorRed,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: Colors.white,
        elevation: 0,
        centerTitle: true,
        iconTheme: IconThemeData(color: ZippyraColors.textPrimary),
        titleTextStyle: TextStyle(
          color: ZippyraColors.textPrimary,
          fontSize: 18,
          fontWeight: FontWeight.w700,
        ),
      ),
      fontFamily: 'Plus Jakarta Sans',
    );
  }
}

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppTheme {
  // Цветовая палитра - современные градиенты
  static const Color primaryColor = Color(0xFF6C5CE7);
  static const Color secondaryColor = Color(0xFFA29BFE);
  static const Color accentColor = Color(0xFFFF6B9D);
  static const Color darkPurple = Color(0xFF5F27CD);
  static const Color lightPurple = Color(0xFFB2A4FF);

  static const Color backgroundColor = Color(0xFF0F0E17);
  static const Color surfaceColor = Color(0xFF1E1D2B);
  static const Color cardColor = Color(0xFF2A2938);

  static const Color textPrimary = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0xFFB8B8D1);
  static const Color texthint = Color(0xFF6E6D7A);

  static const Color errorColor = Color(0xFFFF6B6B);
  static const Color successColor = Color(0xFF51CF66);
  static const Color warningColor = Color(0xFFFFA726);

  // Градиенты
  static const LinearGradient primaryGradient = LinearGradient(
    colors: [primaryColor, accentColor],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  static const LinearGradient secondaryGradient = LinearGradient(
    colors: [darkPurple, primaryColor, accentColor],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  static const LinearGradient backgroundGradient = LinearGradient(
    colors: [Color(0xFF0F0E17), Color(0xFF1A1828), Color(0xFF1E1D2B)],
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
  );

  static const LinearGradient shimmerGradient = LinearGradient(
    colors: [Color(0x00FFFFFF), Color(0x33FFFFFF), Color(0x00FFFFFF)],
  );

  // Тема приложения
  static ThemeData get darkTheme {
    return ThemeData(
      brightness: Brightness.dark,
      primaryColor: primaryColor,
      scaffoldBackgroundColor: backgroundColor,
      colorScheme: const ColorScheme.dark(
        primary: primaryColor,
        secondary: secondaryColor,
        surface: surfaceColor,
        error: errorColor,
      ),
      textTheme: GoogleFonts.poppinsTextTheme(
        const TextTheme(
          displayLarge: TextStyle(
            fontSize: 32,
            fontWeight: FontWeight.bold,
            color: textPrimary,
            letterSpacing: -0.5,
          ),
          displayMedium: TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.bold,
            color: textPrimary,
          ),
          headlineMedium: TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.w600,
            color: textPrimary,
          ),
          titleLarge: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.w600,
            color: textPrimary,
          ),
          bodyLarge: TextStyle(fontSize: 16, color: textPrimary),
          bodyMedium: TextStyle(fontSize: 14, color: textSecondary),
          labelLarge: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: textPrimary,
          ),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: surfaceColor.withValues(alpha: 0.5),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(
            color: texthint.withValues(alpha: 0.2),
            width: 1,
          ),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: primaryColor, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: errorColor, width: 1),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 20,
          vertical: 18,
        ),
      ),
    );
  }

  // Тени
  static List<BoxShadow> get cardShadow => [
    BoxShadow(
      color: primaryColor.withValues(alpha: 0.15),
      blurRadius: 20,
      offset: const Offset(0, 10),
    ),
  ];

  static List<BoxShadow> get glowShadow => [
    BoxShadow(
      color: primaryColor.withValues(alpha: 0.3),
      blurRadius: 30,
      spreadRadius: 5,
    ),
  ];

  static List<BoxShadow> get buttonShadow => [
    BoxShadow(
      color: primaryColor.withValues(alpha: 0.4),
      blurRadius: 15,
      offset: const Offset(0, 8),
    ),
  ];
}

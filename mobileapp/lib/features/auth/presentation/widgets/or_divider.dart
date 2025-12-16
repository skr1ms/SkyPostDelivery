import 'package:flutter/material.dart';
import '../../../../core/theme/app_theme.dart';

class OrDivider extends StatelessWidget {
  const OrDivider({super.key});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Container(
            height: 1,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  Colors.transparent,
                  AppTheme.texthint.withValues(alpha: 0.3),
                ],
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: AppTheme.surfaceColor.withValues(alpha: 0.5),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(
                color: AppTheme.texthint.withValues(alpha: 0.2),
                width: 1,
              ),
            ),
            child: Text(
              'ИЛИ',
              style: TextStyle(
                color: AppTheme.texthint.withValues(alpha: 0.8),
                fontSize: 12,
                fontWeight: FontWeight.w600,
                letterSpacing: 1.2,
              ),
            ),
          ),
        ),
        Expanded(
          child: Container(
            height: 1,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  AppTheme.texthint.withValues(alpha: 0.3),
                  Colors.transparent,
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

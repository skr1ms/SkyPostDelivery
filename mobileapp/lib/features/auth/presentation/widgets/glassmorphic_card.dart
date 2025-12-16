import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../../core/theme/app_theme.dart';

class GlassmorphicCard extends StatelessWidget {
  final Widget child;
  final double blur;
  final double opacity;
  final EdgeInsets? padding;
  final EdgeInsets? margin;
  final BorderRadius? borderRadius;

  const GlassmorphicCard({
    super.key,
    required this.child,
    this.blur = 10,
    this.opacity = 0.1,
    this.padding,
    this.margin,
    this.borderRadius,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: margin,
      decoration: BoxDecoration(
        borderRadius: borderRadius ?? BorderRadius.circular(24),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.1),
          width: 1.5,
        ),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Colors.white.withValues(alpha: opacity),
            Colors.white.withValues(alpha: opacity * 0.5),
          ],
        ),
        boxShadow: [
          BoxShadow(
            color: AppTheme.primaryColor.withValues(alpha: 0.1),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: borderRadius ?? BorderRadius.circular(24),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: blur, sigmaY: blur),
          child: Container(
            padding: padding ?? const EdgeInsets.all(24),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  AppTheme.surfaceColor.withValues(alpha: 0.8),
                  AppTheme.surfaceColor.withValues(alpha: 0.6),
                ],
              ),
            ),
            child: child,
          ),
        ),
      ),
    );
  }
}

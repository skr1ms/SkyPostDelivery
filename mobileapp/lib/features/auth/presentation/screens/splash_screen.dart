import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'dart:math' as math;
import '../../../../core/theme/app_theme.dart';
import '../providers/auth_provider.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen>
    with TickerProviderStateMixin {
  late AnimationController _rotationController;
  late AnimationController _scaleController;
  late AnimationController _fadeController;
  late Animation<double> _scaleAnimation;
  late Animation<double> _fadeAnimation;
  String _statusText = 'Загрузка...';

  @override
  void initState() {
    super.initState();

    _rotationController = AnimationController(
      duration: const Duration(seconds: 3),
      vsync: this,
    )..repeat();

    _scaleController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    );

    _fadeController = AnimationController(
      duration: const Duration(milliseconds: 800),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _scaleController, curve: Curves.elasticOut),
    );

    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(parent: _fadeController, curve: Curves.easeIn));

    _startAnimation();
  }

  void _startAnimation() async {
    await Future.delayed(const Duration(milliseconds: 300));
    _scaleController.forward();

    await Future.delayed(const Duration(milliseconds: 500));
    _fadeController.forward();

    await Future.delayed(const Duration(seconds: 2));

    if (mounted) {
      setState(() {
        _statusText = 'Проверка подключения...';
      });

      final authProvider = Provider.of<AuthProvider>(context, listen: false);

      await authProvider.loadStoredAuth();

      if (!mounted) return;

      if (authProvider.isOfflineMode) {
        setState(() {
          _statusText = 'Оффлайн-режим';
        });
        await Future.delayed(const Duration(milliseconds: 500));
      }

      if (!mounted) return;

      if (authProvider.isAuthenticated) {
        Navigator.of(context).pushReplacementNamed('/main');
      } else {
        Navigator.of(context).pushReplacementNamed('/auth');
      }
    }
  }

  @override
  void dispose() {
    _rotationController.dispose();
    _scaleController.dispose();
    _fadeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: SystemUiOverlayStyle.light.copyWith(
        statusBarColor: Colors.transparent,
        statusBarIconBrightness: Brightness.light,
      ),
      child: Scaffold(
        body: Container(
          decoration: const BoxDecoration(
            gradient: AppTheme.backgroundGradient,
          ),
          child: Stack(
            children: [
              // Animated gradient circles
              AnimatedBuilder(
                animation: _rotationController,
                builder: (context, child) {
                  return CustomPaint(
                    painter: SplashCirclesPainter(
                      animation: _rotationController.value,
                    ),
                    child: Container(),
                  );
                },
              ),

              // Content
              Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    // Animated Logo
                    AnimatedBuilder(
                      animation: _scaleAnimation,
                      builder: (context, child) {
                        return Transform.scale(
                          scale: _scaleAnimation.value,
                          child: Container(
                            width: 140,
                            height: 140,
                            decoration: BoxDecoration(
                              gradient: AppTheme.primaryGradient,
                              borderRadius: BorderRadius.circular(35),
                              boxShadow: [
                                BoxShadow(
                                  color: AppTheme.primaryColor.withValues(
                                    alpha: 0.5,
                                  ),
                                  blurRadius: 40,
                                  spreadRadius: 10,
                                ),
                              ],
                            ),
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(35),
                              child: Image.asset(
                                'assets/icon/logo.jpg',
                                width: 140,
                                height: 140,
                                fit: BoxFit.cover,
                              ),
                            ),
                          ),
                        );
                      },
                    ),

                    const SizedBox(height: 40),

                    // App Name
                    FadeTransition(
                      opacity: _fadeAnimation,
                      child: Column(
                        children: [
                          ShaderMask(
                            shaderCallback: (bounds) =>
                                AppTheme.primaryGradient.createShader(bounds),
                            child: const Text(
                              'SkyPost Delivery',
                              style: TextStyle(
                                fontSize: 36,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                                letterSpacing: 1,
                              ),
                            ),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            'Доставка будущего',
                            style: TextStyle(
                              fontSize: 16,
                              color: AppTheme.textSecondary.withValues(
                                alpha: 0.8,
                              ),
                              letterSpacing: 2,
                              fontWeight: FontWeight.w300,
                            ),
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 60),

                    // Loading indicator
                    FadeTransition(
                      opacity: _fadeAnimation,
                      child: Column(
                        children: [
                          SizedBox(
                            width: 40,
                            height: 40,
                            child: CircularProgressIndicator(
                              strokeWidth: 3,
                              valueColor: AlwaysStoppedAnimation<Color>(
                                AppTheme.primaryColor.withValues(alpha: 0.8),
                              ),
                            ),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _statusText,
                            style: TextStyle(
                              fontSize: 14,
                              color: AppTheme.textSecondary.withValues(
                                alpha: 0.8,
                              ),
                              fontWeight: FontWeight.w300,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class SplashCirclesPainter extends CustomPainter {
  final double animation;

  SplashCirclesPainter({required this.animation});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()..style = PaintingStyle.fill;

    // Circle 1
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.primaryColor.withValues(alpha: 0.3),
            AppTheme.primaryColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.2 + math.sin(animation * 2 * math.pi) * 80,
              size.height * 0.3 + math.cos(animation * 2 * math.pi) * 80,
            ),
            radius: 200,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.2 + math.sin(animation * 2 * math.pi) * 80,
        size.height * 0.3 + math.cos(animation * 2 * math.pi) * 80,
      ),
      200,
      paint,
    );

    // Circle 2
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.accentColor.withValues(alpha: 0.25),
            AppTheme.accentColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.8 + math.cos(animation * 2 * math.pi + 1) * 90,
              size.height * 0.7 + math.sin(animation * 2 * math.pi + 1) * 90,
            ),
            radius: 220,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.8 + math.cos(animation * 2 * math.pi + 1) * 90,
        size.height * 0.7 + math.sin(animation * 2 * math.pi + 1) * 90,
      ),
      220,
      paint,
    );

    // Circle 3
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.secondaryColor.withValues(alpha: 0.3),
            AppTheme.secondaryColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.5 + math.sin(animation * 2 * math.pi + 2) * 60,
              size.height * 0.5 + math.cos(animation * 2 * math.pi + 2) * 60,
            ),
            radius: 150,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.5 + math.sin(animation * 2 * math.pi + 2) * 60,
        size.height * 0.5 + math.cos(animation * 2 * math.pi + 2) * 60,
      ),
      150,
      paint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}

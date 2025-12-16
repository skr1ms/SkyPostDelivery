import 'dart:math';
import 'package:flutter/material.dart';
import '../../../../core/theme/app_theme.dart';

class AnimatedBackground extends StatefulWidget {
  final Widget child;

  const AnimatedBackground({super.key, required this.child});

  @override
  State<AnimatedBackground> createState() => _AnimatedBackgroundState();
}

class _AnimatedBackgroundState extends State<AnimatedBackground>
    with TickerProviderStateMixin {
  late AnimationController _controller;
  late AnimationController _floatingController;
  final List<FloatingParticle> particles = [];

  @override
  void initState() {
    super.initState();

    _controller = AnimationController(
      duration: const Duration(seconds: 20),
      vsync: this,
    )..repeat();

    _floatingController = AnimationController(
      duration: const Duration(seconds: 3),
      vsync: this,
    )..repeat(reverse: true);

    // Создаём плавающие частицы
    for (int i = 0; i < 15; i++) {
      particles.add(FloatingParticle());
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    _floatingController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.backgroundGradient),
      child: Stack(
        children: [
          // Animated gradient circles
          AnimatedBuilder(
            animation: _controller,
            builder: (context, child) {
              return CustomPaint(
                painter: GradientCirclesPainter(animation: _controller.value),
                child: Container(),
              );
            },
          ),

          // Floating particles
          ...particles.asMap().entries.map((entry) {
            return AnimatedBuilder(
              animation: _floatingController,
              builder: (context, child) {
                return Positioned(
                  left: entry.value.x,
                  top:
                      entry.value.y +
                      (sin(_floatingController.value * 2 * pi + entry.key) *
                          20),
                  child: Opacity(
                    opacity: 0.3,
                    child: Container(
                      width: entry.value.size,
                      height: entry.value.size,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: RadialGradient(
                          colors: [
                            entry.value.color.withValues(alpha: 0.6),
                            entry.value.color.withValues(alpha: 0),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              },
            );
          }),

          // Content
          widget.child,
        ],
      ),
    );
  }
}

class FloatingParticle {
  late double x;
  late double y;
  late double size;
  late Color color;

  FloatingParticle() {
    final random = Random();
    x = random.nextDouble() * 400;
    y = random.nextDouble() * 800;
    size = random.nextDouble() * 100 + 50;

    final colors = [
      AppTheme.primaryColor,
      AppTheme.accentColor,
      AppTheme.secondaryColor,
      AppTheme.lightPurple,
    ];
    color = colors[random.nextInt(colors.length)];
  }
}

class GradientCirclesPainter extends CustomPainter {
  final double animation;

  GradientCirclesPainter({required this.animation});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()..style = PaintingStyle.fill;

    // Большой круг 1
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.primaryColor.withValues(alpha: 0.2),
            AppTheme.primaryColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.2 + sin(animation * 2 * pi) * 50,
              size.height * 0.3 + cos(animation * 2 * pi) * 50,
            ),
            radius: 150,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.2 + sin(animation * 2 * pi) * 50,
        size.height * 0.3 + cos(animation * 2 * pi) * 50,
      ),
      150,
      paint,
    );

    // Большой круг 2
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.accentColor.withValues(alpha: 0.15),
            AppTheme.accentColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.8 + cos(animation * 2 * pi + 1) * 60,
              size.height * 0.7 + sin(animation * 2 * pi + 1) * 60,
            ),
            radius: 180,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.8 + cos(animation * 2 * pi + 1) * 60,
        size.height * 0.7 + sin(animation * 2 * pi + 1) * 60,
      ),
      180,
      paint,
    );

    // Средний круг
    paint.shader =
        RadialGradient(
          colors: [
            AppTheme.secondaryColor.withValues(alpha: 0.2),
            AppTheme.secondaryColor.withValues(alpha: 0),
          ],
        ).createShader(
          Rect.fromCircle(
            center: Offset(
              size.width * 0.5 + sin(animation * 2 * pi + 2) * 40,
              size.height * 0.5 + cos(animation * 2 * pi + 2) * 40,
            ),
            radius: 120,
          ),
        );

    canvas.drawCircle(
      Offset(
        size.width * 0.5 + sin(animation * 2 * pi + 2) * 40,
        size.height * 0.5 + cos(animation * 2 * pi + 2) * 40,
      ),
      120,
      paint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}

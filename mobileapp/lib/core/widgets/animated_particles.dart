import 'dart:math';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class ParticlesPainter extends CustomPainter {
  final List<Particle> particles;
  final double animationValue;

  ParticlesPainter({required this.particles, required this.animationValue});

  @override
  void paint(Canvas canvas, Size size) {
    for (var particle in particles) {
      final paint = Paint()
        ..color = particle.color.withValues(alpha: particle.opacity)
        ..style = PaintingStyle.fill;

      // Update particle position
      particle.x += particle.velocityX;
      particle.y += particle.velocityY;

      // Wrap around screen
      if (particle.x < 0) particle.x = size.width;
      if (particle.x > size.width) particle.x = 0;
      if (particle.y < 0) particle.y = size.height;
      if (particle.y > size.height) particle.y = 0;

      // Draw particle with glow
      canvas.drawCircle(Offset(particle.x, particle.y), particle.size, paint);

      // Draw glow effect
      final glowPaint = Paint()
        ..shader =
            RadialGradient(
              colors: [
                particle.color.withValues(alpha: particle.opacity * 0.5),
                particle.color.withValues(alpha: 0),
              ],
            ).createShader(
              Rect.fromCircle(
                center: Offset(particle.x, particle.y),
                radius: particle.size * 3,
              ),
            );

      canvas.drawCircle(
        Offset(particle.x, particle.y),
        particle.size * 3,
        glowPaint,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}

class Particle {
  double x;
  double y;
  double size;
  double velocityX;
  double velocityY;
  Color color;
  double opacity;

  Particle({
    required this.x,
    required this.y,
    required this.size,
    required this.velocityX,
    required this.velocityY,
    required this.color,
    required this.opacity,
  });

  factory Particle.random(Size screenSize) {
    final random = Random();
    final colors = [
      AppTheme.primaryColor,
      AppTheme.accentColor,
      AppTheme.secondaryColor,
      AppTheme.lightPurple,
    ];

    return Particle(
      x: random.nextDouble() * screenSize.width,
      y: random.nextDouble() * screenSize.height,
      size: random.nextDouble() * 3 + 1,
      velocityX: (random.nextDouble() - 0.5) * 0.5,
      velocityY: (random.nextDouble() - 0.5) * 0.5,
      color: colors[random.nextInt(colors.length)],
      opacity: random.nextDouble() * 0.6 + 0.2,
    );
  }
}

class AnimatedParticles extends StatefulWidget {
  final int particleCount;
  final Widget child;

  const AnimatedParticles({
    super.key,
    this.particleCount = 30,
    required this.child,
  });

  @override
  State<AnimatedParticles> createState() => _AnimatedParticlesState();
}

class _AnimatedParticlesState extends State<AnimatedParticles>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late List<Particle> particles;

  @override
  void initState() {
    super.initState();

    _controller = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 60),
    )..repeat();

    particles = [];
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (particles.isEmpty) {
      final size = MediaQuery.of(context).size;
      particles = List.generate(
        widget.particleCount,
        (index) => Particle.random(size),
      );
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        AnimatedBuilder(
          animation: _controller,
          builder: (context, child) {
            return CustomPaint(
              painter: ParticlesPainter(
                particles: particles,
                animationValue: _controller.value,
              ),
              size: Size.infinite,
            );
          },
        ),
        widget.child,
      ],
    );
  }
}

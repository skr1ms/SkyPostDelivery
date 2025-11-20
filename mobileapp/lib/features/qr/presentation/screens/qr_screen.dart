import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../auth/presentation/widgets/animated_background.dart';
import '../../../auth/presentation/widgets/glassmorphic_card.dart';

class QRScreen extends StatefulWidget {
  const QRScreen({super.key});

  @override
  State<QRScreen> createState() => _QRScreenState();
}

class _QRScreenState extends State<QRScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 2),
    )..repeat(reverse: true);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_isLoading) {
      _loadQR();
    }
  }

  Future<void> _loadQR() async {
    setState(() => _isLoading = true);

    final authProvider = context.read<AuthProvider>();
    try {
      await authProvider.loadMyQR();
    } catch (e) {
      await authProvider.loadQRFromLocalStorage();
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return Scaffold(
        body: AnimatedBackground(
          child: SafeArea(
            child: const Center(child: CircularProgressIndicator()),
          ),
        ),
      );
    }

    return Scaffold(
      body: AnimatedBackground(
        child: SafeArea(
          child: Consumer<AuthProvider>(
            builder: (context, authProvider, _) {
              final user = authProvider.user;
              final qrCode = authProvider.qrCode;
              final isExpired = authProvider.isQRExpired;
              final isOffline = authProvider.isOfflineMode;

              if (user == null) {
                return const Center(child: Text('Пользователь не найден'));
              }

              if (qrCode == null || qrCode.isEmpty) {
                return const Center(child: Text('QR-код не найден'));
              }

              return SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: Column(
                  children: [
                    // Header
                    Row(
                      children: [
                        const Icon(
                          FontAwesomeIcons.qrcode,
                          color: AppTheme.successColor,
                          size: 28,
                        ),
                        const SizedBox(width: 12),
                        Text(
                          'Мой QR-код',
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                      ],
                    ),

                    const SizedBox(height: 32),

                    if (isOffline)
                      Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: AppTheme.warningColor.withValues(alpha: 0.2),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: AppTheme.warningColor,
                            width: 1,
                          ),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              FontAwesomeIcons.triangleExclamation,
                              color: AppTheme.warningColor,
                              size: 20,
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Оффлайн-режим',
                                    style: Theme.of(context)
                                        .textTheme
                                        .titleSmall
                                        ?.copyWith(
                                          color: AppTheme.warningColor,
                                          fontWeight: FontWeight.bold,
                                        ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'QR-код доступен без интернета',
                                    style: Theme.of(context).textTheme.bodySmall
                                        ?.copyWith(
                                          color: AppTheme.warningColor
                                              .withValues(alpha: 0.9),
                                        ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),

                    // QR Code Card
                    GlassmorphicCard(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        children: [
                          // User Info
                          Text(
                            user.fullName,
                            style: Theme.of(context).textTheme.titleLarge
                                ?.copyWith(fontWeight: FontWeight.bold),
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            user.email ?? user.phoneNumber ?? '',
                            style: Theme.of(context).textTheme.bodyMedium
                                ?.copyWith(color: AppTheme.textSecondary),
                            textAlign: TextAlign.center,
                          ),

                          const SizedBox(height: 32),

                          // QR Code with animation
                          AnimatedBuilder(
                            animation: _pulseController,
                            builder: (context, child) {
                              return Container(
                                decoration: BoxDecoration(
                                  borderRadius: BorderRadius.circular(20),
                                  boxShadow: [
                                    BoxShadow(
                                      color:
                                          (isExpired
                                                  ? AppTheme.errorColor
                                                  : AppTheme.primaryColor)
                                              .withValues(
                                                alpha:
                                                    0.3 +
                                                    _pulseController.value *
                                                        0.3,
                                              ),
                                      blurRadius:
                                          20 + _pulseController.value * 10,
                                      spreadRadius:
                                          2 + _pulseController.value * 2,
                                    ),
                                  ],
                                ),
                                child: ClipRRect(
                                  borderRadius: BorderRadius.circular(20),
                                  child: Container(
                                    color: Colors.white,
                                    padding: const EdgeInsets.all(20),
                                    child: _buildQRImage(qrCode, isExpired),
                                  ),
                                ),
                              );
                            },
                          ),

                          const SizedBox(height: 32),

                          // Status
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                            decoration: BoxDecoration(
                              color:
                                  (isExpired
                                          ? AppTheme.errorColor
                                          : AppTheme.successColor)
                                      .withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: isExpired
                                    ? AppTheme.errorColor
                                    : AppTheme.successColor,
                                width: 1,
                              ),
                            ),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  isExpired
                                      ? FontAwesomeIcons.triangleExclamation
                                      : FontAwesomeIcons.circleCheck,
                                  size: 16,
                                  color: isExpired
                                      ? AppTheme.errorColor
                                      : AppTheme.successColor,
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    isExpired
                                        ? (isOffline
                                              ? 'QR-код истек (работает оффлайн)'
                                              : 'Обновление QR-кода...')
                                        : 'QR-код активен',
                                    style: TextStyle(
                                      color: isExpired
                                          ? AppTheme.errorColor
                                          : AppTheme.successColor,
                                      fontWeight: FontWeight.w600,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                              ],
                            ),
                          ),

                          const SizedBox(height: 16),

                          // Info
                          Text(
                            isOffline
                                ? 'QR-код работает оффлайн. Обновится автоматически при подключении к интернету.'
                                : 'Покажите этот QR-код на постамате для получения заказа',
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(color: AppTheme.textSecondary),
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 24),

                    // Instructions
                    GlassmorphicCard(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(
                                FontAwesomeIcons.circleInfo,
                                color: AppTheme.primaryColor,
                                size: 20,
                              ),
                              const SizedBox(width: 12),
                              Text(
                                'Как получить заказ?',
                                style: Theme.of(context).textTheme.titleMedium
                                    ?.copyWith(fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                          const SizedBox(height: 16),
                          _buildStep('1', 'Подойдите к постамату'),
                          _buildStep('2', 'Покажите QR-код камере'),
                          _buildStep('3', 'Дождитесь открытия ячеек'),
                          _buildStep('4', 'Заберите свои заказы'),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildQRImage(String qrCode, bool isExpired) {
    if (qrCode.isEmpty) {
      return _buildPlaceholderQR(isExpired);
    }

    if (qrCode.startsWith('data:image')) {
      try {
        final base64Data = qrCode.split(',')[1];
        final bytes = base64Decode(base64Data);
        return RepaintBoundary(
          child: Opacity(
            opacity: isExpired ? 0.3 : 1.0,
            child: Image.memory(
              bytes,
              width: 250,
              height: 250,
              fit: BoxFit.contain,
              gaplessPlayback: true,
            ),
          ),
        );
      } catch (e) {
        return _buildPlaceholderQR(isExpired);
      }
    }

    try {
      final bytes = base64Decode(qrCode);
      return RepaintBoundary(
        child: Opacity(
          opacity: isExpired ? 0.3 : 1.0,
          child: Image.memory(
            bytes,
            width: 250,
            height: 250,
            fit: BoxFit.contain,
            gaplessPlayback: true,
            errorBuilder: (context, error, stackTrace) =>
                _buildPlaceholderQR(isExpired),
          ),
        ),
      );
    } catch (e) {
      return _buildPlaceholderQR(isExpired);
    }
  }

  Widget _buildPlaceholderQR(bool isExpired) {
    return Opacity(
      opacity: isExpired ? 0.3 : 1.0,
      child: Container(
        width: 250,
        height: 250,
        decoration: BoxDecoration(
          color: Colors.black,
          borderRadius: BorderRadius.circular(12),
        ),
        child: const Center(
          child: Icon(FontAwesomeIcons.qrcode, size: 100, color: Colors.white),
        ),
      ),
    );
  }

  Widget _buildStep(String number, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              gradient: AppTheme.primaryGradient,
              shape: BoxShape.circle,
            ),
            child: Center(
              child: Text(
                number,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(text, style: Theme.of(context).textTheme.bodyMedium),
          ),
        ],
      ),
    );
  }
}

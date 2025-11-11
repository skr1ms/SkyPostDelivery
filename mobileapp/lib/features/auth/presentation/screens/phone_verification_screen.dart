import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:animate_do/animate_do.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:provider/provider.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/utils/error_handler.dart';
import '../widgets/animated_background.dart';
import '../widgets/glassmorphic_card.dart';
import '../widgets/gradient_button.dart';
import '../providers/auth_provider.dart';

class PhoneVerificationScreen extends StatefulWidget {
  final String phone;
  final String verificationType;

  const PhoneVerificationScreen({
    super.key,
    required this.phone,
    required this.verificationType,
  });

  @override
  State<PhoneVerificationScreen> createState() =>
      _PhoneVerificationScreenState();
}

class _PhoneVerificationScreenState extends State<PhoneVerificationScreen> {
  final List<TextEditingController> _codeControllers = List.generate(
    4,
    (_) => TextEditingController(),
  );
  final List<FocusNode> _focusNodes = List.generate(4, (_) => FocusNode());
  bool _isLoading = false;

  @override
  void dispose() {
    for (var controller in _codeControllers) {
      controller.dispose();
    }
    for (var node in _focusNodes) {
      node.dispose();
    }
    super.dispose();
  }

  String _getCode() {
    return _codeControllers.map((c) => c.text).join();
  }

  void _handleVerification() async {
    final code = _getCode();
    if (code.length != 4) {
      ErrorHandler.showInfo(context, 'Введите 4-значный код');
      return;
    }

    setState(() => _isLoading = true);

    try {
      if (widget.verificationType == 'registration' ||
          widget.verificationType == 'phoneLogin' ||
          widget.verificationType == 'login') {
        final authProvider = Provider.of<AuthProvider>(context, listen: false);
        await authProvider.verifyPhone(widget.phone, code);

        if (mounted) {
          String successMessage;
          if (widget.verificationType == 'registration') {
            successMessage = 'Регистрация завершена!';
          } else if (widget.verificationType == 'phoneLogin' ||
              widget.verificationType == 'login') {
            successMessage = 'Вход выполнен успешно!';
          } else {
            successMessage = 'Телефон успешно подтвержден!';
          }
          ErrorHandler.showSuccess(context, successMessage);
          Navigator.of(
            context,
          ).pushNamedAndRemoveUntil('/main', (route) => false);
        }
      }
    } catch (e) {
      if (mounted) {
        ErrorHandler.showError(context, e);
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  void _resendCode() async {
    try {
      if (widget.verificationType == 'registration') {
        ErrorHandler.showInfo(
          context,
          'Функция повторной отправки в разработке',
        );
      } else if (widget.verificationType == 'phoneLogin' ||
          widget.verificationType == 'login') {
        ErrorHandler.showInfo(context, 'Код отправлен повторно');
      }
    } catch (e) {
      if (mounted) {
        ErrorHandler.showError(context, e);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const FaIcon(FontAwesomeIcons.arrowLeft, size: 20),
          color: AppTheme.textPrimary,
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: AnimatedBackground(
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: EdgeInsets.only(
                left: 24,
                right: 24,
                top: 24,
                bottom: MediaQuery.of(context).viewInsets.bottom + 24,
              ),
              child: FadeIn(
                duration: const Duration(milliseconds: 600),
                child: GlassmorphicCard(
                  child: Padding(
                    padding: const EdgeInsets.all(32),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        FadeInDown(
                          delay: const Duration(milliseconds: 200),
                          child: Container(
                            padding: const EdgeInsets.all(20),
                            decoration: BoxDecoration(
                              gradient: AppTheme.primaryGradient,
                              shape: BoxShape.circle,
                              boxShadow: [
                                BoxShadow(
                                  color: AppTheme.primaryColor.withValues(
                                    alpha: 0.3,
                                  ),
                                  blurRadius: 20,
                                  spreadRadius: 5,
                                ),
                              ],
                            ),
                            child: const FaIcon(
                              FontAwesomeIcons.mobileScreen,
                              size: 40,
                              color: Colors.white,
                            ),
                          ),
                        ),
                        const SizedBox(height: 32),
                        FadeInDown(
                          delay: const Duration(milliseconds: 300),
                          child: const Text(
                            'Подтверждение',
                            style: TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.bold,
                              color: AppTheme.textPrimary,
                            ),
                          ),
                        ),
                        const SizedBox(height: 12),
                        FadeInDown(
                          delay: const Duration(milliseconds: 400),
                          child: Text(
                            'Введите код из SMS\nотправленный на ${widget.phone}',
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 14,
                              color: AppTheme.textSecondary.withValues(
                                alpha: 0.8,
                              ),
                              height: 1.5,
                            ),
                          ),
                        ),
                        const SizedBox(height: 48),
                        FadeInUp(
                          delay: const Duration(milliseconds: 500),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                            children: List.generate(4, (index) {
                              return _buildCodeInput(index);
                            }),
                          ),
                        ),
                        const SizedBox(height: 32),
                        FadeInUp(
                          delay: const Duration(milliseconds: 600),
                          child: GradientButton(
                            text: 'Подтвердить',
                            onPressed: _handleVerification,
                            isLoading: _isLoading,
                            icon: FontAwesomeIcons.check,
                          ),
                        ),
                        const SizedBox(height: 24),
                        FadeInUp(
                          delay: const Duration(milliseconds: 700),
                          child: TextButton(
                            onPressed: _resendCode,
                            child: Text(
                              'Отправить код повторно',
                              style: TextStyle(
                                color: AppTheme.primaryColor,
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCodeInput(int index) {
    return Container(
      width: 60,
      height: 70,
      decoration: BoxDecoration(
        color: AppTheme.surfaceColor.withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: _codeControllers[index].text.isNotEmpty
              ? AppTheme.primaryColor
              : AppTheme.texthint.withValues(alpha: 0.2),
          width: 2,
        ),
        boxShadow: _codeControllers[index].text.isNotEmpty
            ? [
                BoxShadow(
                  color: AppTheme.primaryColor.withValues(alpha: 0.2),
                  blurRadius: 10,
                  spreadRadius: 2,
                ),
              ]
            : null,
      ),
      child: TextField(
        controller: _codeControllers[index],
        focusNode: _focusNodes[index],
        textAlign: TextAlign.center,
        keyboardType: TextInputType.number,
        maxLength: 1,
        style: const TextStyle(
          fontSize: 28,
          fontWeight: FontWeight.bold,
          color: AppTheme.textPrimary,
        ),
        decoration: const InputDecoration(
          counterText: '',
          border: InputBorder.none,
          contentPadding: EdgeInsets.zero,
        ),
        inputFormatters: [FilteringTextInputFormatter.digitsOnly],
        onChanged: (value) {
          setState(() {});
          if (value.isNotEmpty && index < 3) {
            _focusNodes[index + 1].requestFocus();
          } else if (value.isEmpty && index > 0) {
            _focusNodes[index - 1].requestFocus();
          }
          if (index == 3 && value.isNotEmpty) {
            _handleVerification();
          }
        },
      ),
    );
  }
}

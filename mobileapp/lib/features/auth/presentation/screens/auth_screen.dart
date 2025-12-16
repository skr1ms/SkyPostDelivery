import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:animate_do/animate_do.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:provider/provider.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/utils/error_handler.dart';
import '../../../../core/utils/phone_utils.dart';
import '../../../../core/network/http_client.dart';
import '../widgets/animated_background.dart';
import '../widgets/glassmorphic_card.dart';
import '../widgets/custom_text_field.dart';
import '../widgets/gradient_button.dart';
import '../providers/auth_provider.dart';
import 'phone_verification_screen.dart';

class AuthScreen extends StatefulWidget {
  const AuthScreen({super.key});

  @override
  State<AuthScreen> createState() => _AuthScreenState();
}

class _AuthScreenState extends State<AuthScreen>
    with SingleTickerProviderStateMixin {
  final PageController _pageController = PageController();
  late TabController _tabController;

  // Controllers для логина
  final TextEditingController _loginEmailController = TextEditingController();
  final TextEditingController _loginPasswordController =
      TextEditingController();

  // Controllers для регистрации
  final TextEditingController _registerNameController = TextEditingController();
  final TextEditingController _registerEmailController =
      TextEditingController();
  final TextEditingController _registerPhoneController =
      TextEditingController();
  final TextEditingController _registerPasswordController =
      TextEditingController();
  final TextEditingController _registerConfirmPasswordController =
      TextEditingController();

  final GlobalKey<FormState> _loginFormKey = GlobalKey<FormState>();
  final GlobalKey<FormState> _registerFormKey = GlobalKey<FormState>();

  bool _isLoginLoading = false;
  bool _isRegisterLoading = false;
  int _currentIndex = 0;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(() {
      if (_tabController.indexIsChanging) {
        _pageController.animateToPage(
          _tabController.index,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
        setState(() {
          _currentIndex = _tabController.index;
        });
      }
    });
  }

  @override
  void dispose() {
    _pageController.dispose();
    _tabController.dispose();
    _loginEmailController.dispose();
    _loginPasswordController.dispose();
    _registerNameController.dispose();
    _registerEmailController.dispose();
    _registerPhoneController.dispose();
    _registerPasswordController.dispose();
    _registerConfirmPasswordController.dispose();
    super.dispose();
  }

  void _handleLogin() async {
    if (!_loginFormKey.currentState!.validate()) return;

    setState(() => _isLoginLoading = true);

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      await authProvider.login(
        _loginEmailController.text.trim(),
        _loginPasswordController.text,
      );

      if (mounted) {
        ErrorHandler.showSuccess(context, 'Вход выполнен успешно!');
        Navigator.of(context).pushReplacementNamed('/main');
      }
    } catch (e) {
      final errorMessage = e.toString().replaceAll('Exception: ', '');

      if (errorMessage.startsWith('PHONE_NOT_VERIFIED:')) {
        final phone = errorMessage.split(':')[1];

        if (mounted) {
          setState(() => _isLoginLoading = false);

          showDialog(
            context: context,
            barrierDismissible: false,
            builder: (context) => Dialog(
              backgroundColor: Colors.transparent,
              child: FadeInDown(
                duration: const Duration(milliseconds: 400),
                child: GlassmorphicCard(
                  padding: const EdgeInsets.all(28),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            colors: [
                              AppTheme.accentColor,
                              AppTheme.accentColor.withValues(alpha: 0.7),
                            ],
                          ),
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: AppTheme.accentColor.withValues(
                                alpha: 0.3,
                              ),
                              blurRadius: 20,
                              spreadRadius: 5,
                            ),
                          ],
                        ),
                        child: const FaIcon(
                          FontAwesomeIcons.triangleExclamation,
                          size: 32,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'Требуется верификация',
                        style: TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: AppTheme.textPrimary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'Вы не подтвердили номер телефона.\nМы отправили вам SMS с кодом.',
                        style: TextStyle(
                          fontSize: 14,
                          color: AppTheme.textSecondary.withValues(alpha: 0.8),
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      Row(
                        children: [
                          Expanded(
                            child: TextButton(
                              onPressed: () => Navigator.pop(context),
                              style: TextButton.styleFrom(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 16,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  side: BorderSide(
                                    color: AppTheme.texthint.withValues(
                                      alpha: 0.3,
                                    ),
                                  ),
                                ),
                              ),
                              child: Text(
                                'Отмена',
                                style: TextStyle(
                                  color: AppTheme.textSecondary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            flex: 2,
                            child: GradientButton(
                              text: 'Подтвердить',
                              onPressed: () {
                                Navigator.pop(context);
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (context) =>
                                        PhoneVerificationScreen(
                                          phone: phone,
                                          verificationType: 'login',
                                        ),
                                  ),
                                );
                              },
                              isLoading: false,
                              icon: FontAwesomeIcons.check,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          );
        }
        return;
      }

      if (mounted) {
        ErrorHandler.showError(context, errorMessage);
      }
    } finally {
      if (mounted) {
        setState(() => _isLoginLoading = false);
      }
    }
  }

  void _handleRegister() async {
    if (!_registerFormKey.currentState!.validate()) return;

    setState(() => _isRegisterLoading = true);

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      await authProvider.register(
        fullName: _registerNameController.text.trim(),
        email: _registerEmailController.text.trim(),
        phone: _registerPhoneController.text.trim(),
        password: _registerPasswordController.text,
      );

      if (mounted) {
        ErrorHandler.showSuccess(context, 'Код отправлен на ваш номер');
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => PhoneVerificationScreen(
              phone: _registerPhoneController.text.trim(),
              verificationType: 'registration',
            ),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ErrorHandler.showError(context, e);
      }
    } finally {
      if (mounted) {
        setState(() => _isRegisterLoading = false);
      }
    }
  }

  void _showPhoneLoginDialog() {
    final phoneController = TextEditingController();
    final httpClient = HttpClient();
    bool isLoading = false;

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => Dialog(
          backgroundColor: Colors.transparent,
          insetPadding: const EdgeInsets.symmetric(
            horizontal: 20,
            vertical: 24,
          ),
          child: FadeInDown(
            duration: const Duration(milliseconds: 400),
            child: SingleChildScrollView(
              child: Padding(
                padding: EdgeInsets.only(
                  bottom: MediaQuery.of(context).viewInsets.bottom,
                ),
                child: GlassmorphicCard(
                  padding: const EdgeInsets.all(28),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            colors: [
                              AppTheme.accentColor,
                              AppTheme.accentColor.withValues(alpha: 0.7),
                            ],
                          ),
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: AppTheme.accentColor.withValues(
                                alpha: 0.3,
                              ),
                              blurRadius: 20,
                              spreadRadius: 5,
                            ),
                          ],
                        ),
                        child: const FaIcon(
                          FontAwesomeIcons.mobileScreen,
                          size: 32,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'Вход по телефону',
                        style: TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: AppTheme.textPrimary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Мы отправим вам код подтверждения',
                        style: TextStyle(
                          fontSize: 14,
                          color: AppTheme.textSecondary.withValues(alpha: 0.8),
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      CustomTextField(
                        controller: phoneController,
                        label: 'Номер телефона',
                        hint: '+7 (___) ___-__-__',
                        icon: FontAwesomeIcons.phone,
                        keyboardType: TextInputType.phone,
                      ),
                      const SizedBox(height: 24),
                      Row(
                        children: [
                          Expanded(
                            child: TextButton(
                              onPressed: isLoading
                                  ? null
                                  : () => Navigator.pop(context),
                              style: TextButton.styleFrom(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 16,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  side: BorderSide(
                                    color: AppTheme.texthint.withValues(
                                      alpha: 0.3,
                                    ),
                                  ),
                                ),
                              ),
                              child: Text(
                                'Отмена',
                                style: TextStyle(
                                  color: AppTheme.textSecondary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            flex: 2,
                            child: GradientButton(
                              text: 'Получить код',
                              onPressed: () async {
                                if (isLoading) return;

                                if (phoneController.text.trim().isEmpty) {
                                  ErrorHandler.showInfo(
                                    context,
                                    'Введите номер телефона',
                                  );
                                  return;
                                }

                                setState(() => isLoading = true);
                                try {
                                  final phone = PhoneUtils.cleanPhone(
                                    phoneController.text,
                                  );
                                  await httpClient.post(
                                    '/auth/login/phone',
                                    body: {'phone': phone},
                                  );

                                  if (context.mounted) {
                                    Navigator.pop(context);
                                    Navigator.of(context).push(
                                      MaterialPageRoute(
                                        builder: (context) =>
                                            PhoneVerificationScreen(
                                              phone: phone,
                                              verificationType: 'phoneLogin',
                                            ),
                                      ),
                                    );
                                  }
                                } catch (e) {
                                  setState(() => isLoading = false);
                                  if (context.mounted) {
                                    ErrorHandler.showError(context, e);
                                  }
                                }
                              },
                              isLoading: isLoading,
                              icon: FontAwesomeIcons.paperPlane,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _showPasswordResetDialog() {
    final phoneController = TextEditingController();
    final codeController = TextEditingController();
    final newPasswordController = TextEditingController();
    final httpClient = HttpClient();
    bool isCodeSent = false;
    bool isLoading = false;

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => Dialog(
          backgroundColor: Colors.transparent,
          insetPadding: const EdgeInsets.symmetric(
            horizontal: 20,
            vertical: 24,
          ),
          child: FadeInDown(
            duration: const Duration(milliseconds: 400),
            child: SingleChildScrollView(
              child: Padding(
                padding: EdgeInsets.only(
                  bottom: MediaQuery.of(context).viewInsets.bottom,
                ),
                child: GlassmorphicCard(
                  padding: const EdgeInsets.all(28),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        padding: const EdgeInsets.all(16),
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
                          FontAwesomeIcons.key,
                          size: 32,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'Восстановление пароля',
                        style: TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: AppTheme.textPrimary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        isCodeSent
                            ? 'Введите код и новый пароль'
                            : 'Введите номер телефона',
                        style: TextStyle(
                          fontSize: 14,
                          color: AppTheme.textSecondary.withValues(alpha: 0.8),
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      CustomTextField(
                        controller: phoneController,
                        label: 'Номер телефона',
                        hint: '+7 (___) ___-__-__',
                        icon: FontAwesomeIcons.phone,
                        keyboardType: TextInputType.phone,
                      ),
                      if (isCodeSent) ...[
                        const SizedBox(height: 16),
                        CustomTextField(
                          controller: codeController,
                          label: 'Код подтверждения',
                          hint: '____',
                          icon: FontAwesomeIcons.shield,
                          keyboardType: TextInputType.number,
                        ),
                        const SizedBox(height: 16),
                        CustomTextField(
                          controller: newPasswordController,
                          label: 'Новый пароль',
                          hint: 'Минимум 6 символов',
                          icon: FontAwesomeIcons.lock,
                          isPassword: true,
                        ),
                      ],
                      const SizedBox(height: 24),
                      Row(
                        children: [
                          Expanded(
                            child: TextButton(
                              onPressed: isLoading
                                  ? null
                                  : () => Navigator.pop(context),
                              style: TextButton.styleFrom(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 16,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  side: BorderSide(
                                    color: AppTheme.texthint.withValues(
                                      alpha: 0.3,
                                    ),
                                  ),
                                ),
                              ),
                              child: Text(
                                'Отмена',
                                style: TextStyle(
                                  color: AppTheme.textSecondary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            flex: 2,
                            child: GradientButton(
                              text: isCodeSent ? 'Изменить' : 'Получить код',
                              onPressed: () async {
                                if (isLoading) return;

                                if (!isCodeSent) {
                                  if (phoneController.text.trim().isEmpty) {
                                    ErrorHandler.showInfo(
                                      context,
                                      'Введите номер телефона',
                                    );
                                    return;
                                  }

                                  setState(() => isLoading = true);
                                  try {
                                    final phone = PhoneUtils.cleanPhone(
                                      phoneController.text,
                                    );
                                    await httpClient.post(
                                      '/auth/password/reset/request',
                                      body: {'phone': phone},
                                    );
                                    setState(() {
                                      isCodeSent = true;
                                      isLoading = false;
                                    });
                                    if (context.mounted) {
                                      ErrorHandler.showSuccess(
                                        context,
                                        'Код отправлен 📱',
                                      );
                                    }
                                  } catch (e) {
                                    setState(() => isLoading = false);
                                    if (context.mounted) {
                                      ErrorHandler.showError(context, e);
                                    }
                                  }
                                } else {
                                  if (codeController.text.trim().isEmpty ||
                                      newPasswordController.text.isEmpty) {
                                    ErrorHandler.showInfo(
                                      context,
                                      'Заполните все поля',
                                    );
                                    return;
                                  }

                                  if (newPasswordController.text.length < 6) {
                                    ErrorHandler.showInfo(
                                      context,
                                      'Пароль минимум 6 символов',
                                    );
                                    return;
                                  }

                                  setState(() => isLoading = true);
                                  try {
                                    final phone = PhoneUtils.cleanPhone(
                                      phoneController.text,
                                    );
                                    await httpClient.post(
                                      '/auth/password/reset',
                                      body: {
                                        'phone': phone,
                                        'code': codeController.text.trim(),
                                        'new_password':
                                            newPasswordController.text,
                                      },
                                    );
                                    if (context.mounted) {
                                      Navigator.pop(context);
                                      ErrorHandler.showSuccess(
                                        context,
                                        'Пароль успешно изменен!',
                                      );
                                    }
                                  } catch (e) {
                                    setState(() => isLoading = false);
                                    if (context.mounted) {
                                      ErrorHandler.showError(context, e);
                                    }
                                  }
                                }
                              },
                              isLoading: isLoading,
                              icon: isCodeSent
                                  ? FontAwesomeIcons.check
                                  : FontAwesomeIcons.paperPlane,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: SystemUiOverlayStyle.light.copyWith(
        statusBarColor: Colors.transparent,
        statusBarIconBrightness: Brightness.light,
      ),
      child: Scaffold(
        resizeToAvoidBottomInset: true,
        body: AnimatedBackground(
          child: SafeArea(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24),
              child: Column(
                children: [
                  const SizedBox(height: 40),
                  // Логотип и заголовок
                  FadeInDown(
                    duration: const Duration(milliseconds: 600),
                    child: Column(
                      children: [
                        // Логотип с градиентом и свечением
                        Container(
                          width: 80,
                          height: 80,
                          decoration: BoxDecoration(
                            gradient: AppTheme.primaryGradient,
                            borderRadius: BorderRadius.circular(24),
                            boxShadow: AppTheme.glowShadow,
                          ),
                          child: ClipRRect(
                            borderRadius: BorderRadius.circular(24),
                            child: Image.asset(
                              'assets/icon/logo.jpg',
                              width: 80,
                              height: 80,
                              fit: BoxFit.cover,
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        ShaderMask(
                          shaderCallback: (bounds) =>
                              AppTheme.primaryGradient.createShader(bounds),
                          child: const Text(
                            'SkyPost Delivery',
                            style: TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Доставка будущего',
                          style: TextStyle(
                            fontSize: 14,
                            color: AppTheme.textSecondary.withValues(
                              alpha: 0.8,
                            ),
                            letterSpacing: 0.3,
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 40),

                  // Переключатель Вход/Регистрация
                  FadeInDown(
                    delay: const Duration(milliseconds: 200),
                    duration: const Duration(milliseconds: 600),
                    child: _buildTabSelector(),
                  ),

                  const SizedBox(height: 32),

                  // Формы
                  Expanded(
                    child: PageView(
                      controller: _pageController,
                      onPageChanged: (index) {
                        _tabController.animateTo(index);
                        setState(() => _currentIndex = index);
                      },
                      children: [_buildLoginForm(), _buildRegisterForm()],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTabSelector() {
    return GlassmorphicCard(
      padding: const EdgeInsets.all(6),
      child: Row(
        children: [
          Expanded(child: _buildTab('Вход', 0)),
          const SizedBox(width: 6),
          Expanded(child: _buildTab('Регистрация', 1)),
        ],
      ),
    );
  }

  Widget _buildTab(String text, int index) {
    final isSelected = _currentIndex == index;

    return GestureDetector(
      onTap: () {
        _tabController.animateTo(index);
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          gradient: isSelected ? AppTheme.primaryGradient : null,
          borderRadius: BorderRadius.circular(14),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: AppTheme.primaryColor.withValues(alpha: 0.4),
                    blurRadius: 12,
                    offset: const Offset(0, 4),
                  ),
                ]
              : [],
        ),
        child: Text(
          text,
          textAlign: TextAlign.center,
          style: TextStyle(
            color: isSelected ? Colors.white : AppTheme.textSecondary,
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.3,
          ),
        ),
      ),
    );
  }

  Widget _buildLoginForm() {
    return FadeInUp(
      delay: const Duration(milliseconds: 400),
      duration: const Duration(milliseconds: 600),
      child: SingleChildScrollView(
        child: Form(
          key: _loginFormKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Email поле
              CustomTextField(
                controller: _loginEmailController,
                label: 'Email',
                hint: 'example@mail.com',
                icon: FontAwesomeIcons.envelope,
                keyboardType: TextInputType.emailAddress,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите email';
                  }
                  if (!RegExp(
                    r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$',
                  ).hasMatch(value)) {
                    return 'Введите корректный email';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 20),

              // Password поле
              CustomTextField(
                controller: _loginPasswordController,
                label: 'Пароль',
                hint: '••••••••',
                icon: FontAwesomeIcons.lock,
                isPassword: true,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите пароль';
                  }
                  if (value.length < 6) {
                    return 'Пароль должен быть минимум 6 символов';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 16),

              // Забыли пароль и Войти по телефону
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  TextButton(
                    onPressed: _showPhoneLoginDialog,
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const FaIcon(
                          FontAwesomeIcons.mobileScreen,
                          size: 14,
                          color: AppTheme.accentColor,
                        ),
                        const SizedBox(width: 6),
                        ShaderMask(
                          shaderCallback: (bounds) => LinearGradient(
                            colors: [
                              AppTheme.accentColor,
                              AppTheme.accentColor.withValues(alpha: 0.7),
                            ],
                          ).createShader(bounds),
                          child: const Text(
                            'Войти по телефону',
                            style: TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.w600,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  TextButton(
                    onPressed: _showPasswordResetDialog,
                    child: ShaderMask(
                      shaderCallback: (bounds) =>
                          AppTheme.primaryGradient.createShader(bounds),
                      child: const Text(
                        'Забыли пароль?',
                        style: TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.w600,
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 32),

              // Кнопка входа
              GradientButton(
                text: 'Войти',
                onPressed: _handleLogin,
                isLoading: _isLoginLoading,
                icon: FontAwesomeIcons.arrowRight,
              ),

              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildRegisterForm() {
    return FadeInUp(
      delay: const Duration(milliseconds: 400),
      duration: const Duration(milliseconds: 600),
      child: SingleChildScrollView(
        child: Form(
          key: _registerFormKey,
          child: Column(
            children: [
              // Имя
              CustomTextField(
                controller: _registerNameController,
                label: 'Фамилия Имя',
                hint: 'Иванов Иван',
                icon: FontAwesomeIcons.user,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите Фамилию и Имя';
                  }
                  if (value.split(' ').length < 2) {
                    return 'Введите Фамилию и Имя';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 20),

              // Email
              CustomTextField(
                controller: _registerEmailController,
                label: 'Email',
                hint: 'example@mail.com',
                icon: FontAwesomeIcons.envelope,
                keyboardType: TextInputType.emailAddress,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите email';
                  }
                  if (!RegExp(
                    r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$',
                  ).hasMatch(value)) {
                    return 'Введите корректный email';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 20),

              // Телефон
              CustomTextField(
                controller: _registerPhoneController,
                label: 'Телефон',
                hint: '+7 (___) ___-__-__',
                icon: FontAwesomeIcons.phone,
                keyboardType: TextInputType.phone,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите телефон';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 20),

              // Пароль
              CustomTextField(
                controller: _registerPasswordController,
                label: 'Пароль',
                hint: '••••••••',
                icon: FontAwesomeIcons.lock,
                isPassword: true,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Введите пароль';
                  }
                  if (value.length < 6) {
                    return 'Пароль должен быть минимум 6 символов';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 20),

              // Подтверждение пароля
              CustomTextField(
                controller: _registerConfirmPasswordController,
                label: 'Подтвердите пароль',
                hint: '••••••••',
                icon: FontAwesomeIcons.lock,
                isPassword: true,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Подтвердите пароль';
                  }
                  if (value != _registerPasswordController.text) {
                    return 'Пароли не совпадают';
                  }
                  return null;
                },
              ),

              const SizedBox(height: 32),

              // Кнопка регистрации
              GradientButton(
                text: 'Зарегистрироваться',
                onPressed: _handleRegister,
                isLoading: _isRegisterLoading,
                icon: FontAwesomeIcons.userPlus,
              ),

              const SizedBox(height: 24),

              // Условия использования
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Text.rich(
                  TextSpan(
                    text: 'Регистрируясь, вы принимаете наши ',
                    style: TextStyle(
                      color: AppTheme.textSecondary.withValues(alpha: 0.8),
                      fontSize: 12,
                    ),
                    children: [
                      TextSpan(
                        text: 'Условия использования',
                        style: const TextStyle(
                          color: AppTheme.primaryColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const TextSpan(text: ' и '),
                      TextSpan(
                        text: 'Политику конфиденциальности',
                        style: const TextStyle(
                          color: AppTheme.primaryColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                  textAlign: TextAlign.center,
                ),
              ),

              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }
}

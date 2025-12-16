import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class AppException implements Exception {
  final String message;
  final String? technicalDetails;
  final AppExceptionType type;

  AppException({
    required this.message,
    this.technicalDetails,
    this.type = AppExceptionType.unknown,
  });

  @override
  String toString() => message;
}

enum AppExceptionType {
  network,
  authentication,
  validation,
  notFound,
  serverError,
  unknown,
}

class ErrorHandler {
  static String getUserFriendlyMessage(dynamic error) {
    if (error is AppException) {
      return error.message;
    }

    final errorString = error.toString().toLowerCase();

    if (errorString.contains('serverexception')) {
      if (errorString.contains('invalid credentials')) {
        return 'Неверный логин или пароль';
      }
      if (errorString.contains('user not found')) {
        return 'Пользователь с таким номером не зарегистрирован.\nСначала создайте аккаунт';
      }
      if (errorString.contains('already exists')) {
        return 'Пользователь с такими данными уже существует';
      }
      if (errorString.contains('failed to send sms')) {
        return 'Не удалось отправить SMS. Попробуйте позже';
      }
      if (errorString.contains('phone not verified')) {
        return 'Номер телефона не подтвержден';
      }
      if (errorString.contains('invalid') || errorString.contains('wrong')) {
        return 'Неверные данные';
      }
      return 'Ошибка сервера. Попробуйте позже';
    }

    if (errorString.contains('authexception')) {
      if (errorString.contains('unauthorized')) {
        return 'Необходима авторизация';
      }
      return 'Ошибка авторизации';
    }

    if (errorString.contains('networkexception')) {
      if (errorString.contains('socket') ||
          errorString.contains('failed host lookup') ||
          errorString.contains('network is unreachable')) {
        return 'Проблема с подключением к интернету';
      }
      return 'Ошибка сети. Проверьте подключение';
    }

    if (errorString.contains('timeout')) {
      return 'Превышено время ожидания. Попробуйте позже';
    }

    if (errorString.contains('validation')) {
      return 'Проверьте правильность введенных данных';
    }

    if (errorString.contains('not found') || errorString.contains('404')) {
      return 'Запрашиваемые данные не найдены';
    }

    return 'Произошла ошибка. Попробуйте еще раз';
  }

  static void showError(
    BuildContext context,
    dynamic error, {
    VoidCallback? onRetry,
  }) {
    if (!context.mounted) return;

    final message = getUserFriendlyMessage(error);

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.errorColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        action: onRetry != null
            ? SnackBarAction(
                label: 'Повторить',
                textColor: Colors.white,
                onPressed: onRetry,
              )
            : null,
        duration: const Duration(seconds: 4),
      ),
    );
  }

  static void showSuccess(BuildContext context, String message) {
    if (!context.mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle_outline, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.successColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  static void showInfo(BuildContext context, String message) {
    if (!context.mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.info_outline, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.primaryColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 3),
      ),
    );
  }
}

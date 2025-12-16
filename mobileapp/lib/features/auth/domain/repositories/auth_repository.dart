import '../entities/auth_result.dart';

abstract class AuthRepository {
  Future<String> register({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  });

  Future<AuthResult> login({required String login, required String password});

  Future<String> loginByPhone({required String phone});

  Future<AuthResult> verifyPhone({required String phone, required String code});

  Future<String> requestPasswordReset({required String phone});

  Future<String> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  });

  Future<void> refreshAccessToken(String refreshToken);

  Future<AuthResult> getMe();
}

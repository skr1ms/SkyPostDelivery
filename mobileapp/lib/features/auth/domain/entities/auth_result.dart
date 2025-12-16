import 'user_entity.dart';

class AuthResult {
  final UserEntity user;
  final String? accessToken;
  final String? refreshToken;
  final int? expiresAt;
  final String qrCode;

  const AuthResult({
    required this.user,
    this.accessToken,
    this.refreshToken,
    this.expiresAt,
    required this.qrCode,
  });
}

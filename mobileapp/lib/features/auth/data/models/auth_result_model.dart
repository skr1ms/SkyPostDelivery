import '../../domain/entities/auth_result.dart';
import 'user_model.dart';

class AuthResultModel extends AuthResult {
  const AuthResultModel({
    required super.user,
    super.accessToken,
    super.refreshToken,
    super.expiresAt,
    required super.qrCode,
  });

  factory AuthResultModel.fromJson(Map<String, dynamic> json) {
    return AuthResultModel(
      user: UserModel.fromJson(json['user'] as Map<String, dynamic>),
      accessToken: json['access_token'] as String?,
      refreshToken: json['refresh_token'] as String?,
      expiresAt: json['expires_at'] as int?,
      qrCode: json['qr_code'] as String,
    );
  }
}

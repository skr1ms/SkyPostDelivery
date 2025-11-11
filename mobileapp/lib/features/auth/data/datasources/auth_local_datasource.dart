import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../domain/entities/user_entity.dart';

abstract class AuthLocalDataSource {
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  });

  Future<void> saveQRCode(String qrCode);
  
  Future<void> saveUser(UserEntity user);
  
  Future<void> saveQRExpiresAt(DateTime expiresAt);

  Future<String?> getAccessToken();

  Future<String?> getRefreshToken();

  Future<void> saveDeviceToken(String token);

  Future<String?> getDeviceToken();

  Future<String?> getQRCode();
  
  Future<UserEntity?> getUser();
  
  Future<DateTime?> getQRExpiresAt();

  Future<void> clearAll();
}

class AuthLocalDataSourceImpl implements AuthLocalDataSource {
  final FlutterSecureStorage storage;

  const AuthLocalDataSourceImpl(this.storage);

  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _qrCodeKey = 'qr_code';
  static const String _userDataKey = 'user_data';
  static const String _qrExpiresAtKey = 'qr_expires_at';
  static const String _deviceTokenKey = 'device_token';

  @override
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    await storage.write(key: _accessTokenKey, value: accessToken);
    await storage.write(key: _refreshTokenKey, value: refreshToken);
  }

  @override
  Future<void> saveQRCode(String qrCode) async {
    await storage.write(key: _qrCodeKey, value: qrCode);
  }
  
  @override
  Future<void> saveUser(UserEntity user) async {
    final userData = jsonEncode({
      'id': user.id,
      'fullName': user.fullName,
      'email': user.email,
      'phoneNumber': user.phoneNumber,
      'phoneVerified': user.phoneVerified,
      'role': user.role,
      'createdAt': user.createdAt.toIso8601String(),
      'qrIssuedAt': user.qrIssuedAt?.toIso8601String(),
      'qrExpiresAt': user.qrExpiresAt?.toIso8601String(),
    });
    await storage.write(key: _userDataKey, value: userData);
  }
  
  @override
  Future<void> saveQRExpiresAt(DateTime expiresAt) async {
    await storage.write(key: _qrExpiresAtKey, value: expiresAt.toIso8601String());
  }

  @override
  Future<String?> getAccessToken() async {
    return await storage.read(key: _accessTokenKey);
  }

  @override
  Future<String?> getRefreshToken() async {
    return await storage.read(key: _refreshTokenKey);
  }

  @override
  Future<void> saveDeviceToken(String token) async {
    await storage.write(key: _deviceTokenKey, value: token);
  }

  @override
  Future<String?> getDeviceToken() async {
    return await storage.read(key: _deviceTokenKey);
  }

  @override
  Future<String?> getQRCode() async {
    return await storage.read(key: _qrCodeKey);
  }
  
  @override
  Future<UserEntity?> getUser() async {
    final userDataJson = await storage.read(key: _userDataKey);
    if (userDataJson == null) return null;
    
    try {
      final userData = jsonDecode(userDataJson) as Map<String, dynamic>;
      return UserEntity(
        id: userData['id'] as String,
        fullName: userData['fullName'] as String,
        email: userData['email'] as String?,
        phoneNumber: userData['phoneNumber'] as String?,
        phoneVerified: userData['phoneVerified'] as bool,
        role: userData['role'] as String,
        createdAt: DateTime.parse(userData['createdAt'] as String),
        qrIssuedAt: userData['qrIssuedAt'] != null 
            ? DateTime.parse(userData['qrIssuedAt'] as String)
            : null,
        qrExpiresAt: userData['qrExpiresAt'] != null
            ? DateTime.parse(userData['qrExpiresAt'] as String)
            : null,
      );
    } catch (e) {
      return null;
    }
  }
  
  @override
  Future<DateTime?> getQRExpiresAt() async {
    final expiresAtStr = await storage.read(key: _qrExpiresAtKey);
    if (expiresAtStr == null) return null;
    
    try {
      return DateTime.parse(expiresAtStr);
    } catch (e) {
      return null;
    }
  }

  @override
  Future<void> clearAll() async {
    await storage.delete(key: _accessTokenKey);
    await storage.delete(key: _refreshTokenKey);
    await storage.delete(key: _qrCodeKey);
    await storage.delete(key: _userDataKey);
    await storage.delete(key: _qrExpiresAtKey);
    await storage.delete(key: _deviceTokenKey);
  }
}


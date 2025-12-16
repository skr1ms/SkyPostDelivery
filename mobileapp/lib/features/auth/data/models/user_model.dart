import '../../domain/entities/user_entity.dart';

class UserModel extends UserEntity {
  const UserModel({
    required super.id,
    required super.fullName,
    super.email,
    super.phoneNumber,
    required super.phoneVerified,
    required super.role,
    required super.createdAt,
    super.qrIssuedAt,
    super.qrExpiresAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] as String,
      fullName: json['full_name'] as String,
      email: json['email'] as String?,
      phoneNumber: json['phone_number'] as String?,
      phoneVerified: json['phone_verified'] as bool,
      role: json['role'] as String,
      createdAt: json['created_at'] != null && json['created_at'] != ''
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
      qrIssuedAt: json['qr_issued_at'] != null && json['qr_issued_at'] != ''
          ? DateTime.parse(json['qr_issued_at'] as String)
          : null,
      qrExpiresAt: json['qr_expires_at'] != null && json['qr_expires_at'] != ''
          ? DateTime.parse(json['qr_expires_at'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'full_name': fullName,
      'email': email,
      'phone_number': phoneNumber,
      'phone_verified': phoneVerified,
      'role': role,
      'created_at': createdAt.toIso8601String(),
      if (qrIssuedAt != null) 'qr_issued_at': qrIssuedAt!.toIso8601String(),
      if (qrExpiresAt != null) 'qr_expires_at': qrExpiresAt!.toIso8601String(),
    };
  }

  UserEntity toEntity() {
    return UserEntity(
      id: id,
      fullName: fullName,
      email: email,
      phoneNumber: phoneNumber,
      phoneVerified: phoneVerified,
      role: role,
      createdAt: createdAt,
      qrIssuedAt: qrIssuedAt,
      qrExpiresAt: qrExpiresAt,
    );
  }
}


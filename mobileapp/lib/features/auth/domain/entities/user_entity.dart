class UserEntity {
  final String id;
  final String fullName;
  final String? email;
  final String? phoneNumber;
  final bool phoneVerified;
  final String role;
  final DateTime createdAt;
  final DateTime? qrIssuedAt;
  final DateTime? qrExpiresAt;

  const UserEntity({
    required this.id,
    required this.fullName,
    this.email,
    this.phoneNumber,
    required this.phoneVerified,
    required this.role,
    required this.createdAt,
    this.qrIssuedAt,
    this.qrExpiresAt,
  });
}


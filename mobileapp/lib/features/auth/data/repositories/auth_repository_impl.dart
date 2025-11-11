import '../../domain/entities/auth_result.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';
import '../datasources/auth_local_datasource.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthRemoteDataSource remoteDataSource;
  final AuthLocalDataSource localDataSource;

  const AuthRepositoryImpl({
    required this.remoteDataSource,
    required this.localDataSource,
  });

  @override
  Future<String> register({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  }) {
    return remoteDataSource.register(
      fullName: fullName,
      email: email,
      phone: phone,
      password: password,
    );
  }

  @override
  Future<AuthResult> login({
    required String login,
    required String password,
  }) async {
    final result = await remoteDataSource.login(
      login: login,
      password: password,
    );

    if (result.accessToken != null && result.refreshToken != null) {
      await localDataSource.saveTokens(
        accessToken: result.accessToken!,
        refreshToken: result.refreshToken!,
      );
    }
    await localDataSource.saveQRCode(result.qrCode);

    return result;
  }

  @override
  Future<String> loginByPhone({required String phone}) {
    return remoteDataSource.loginByPhone(phone: phone);
  }

  @override
  Future<AuthResult> verifyPhone({
    required String phone,
    required String code,
  }) async {
    final result = await remoteDataSource.verifyPhone(phone: phone, code: code);

    if (result.accessToken != null && result.refreshToken != null) {
      await localDataSource.saveTokens(
        accessToken: result.accessToken!,
        refreshToken: result.refreshToken!,
      );
    }
    await localDataSource.saveQRCode(result.qrCode);

    return result;
  }

  @override
  Future<String> requestPasswordReset({required String phone}) {
    return remoteDataSource.requestPasswordReset(phone: phone);
  }

  @override
  Future<String> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  }) {
    return remoteDataSource.resetPassword(
      phone: phone,
      code: code,
      newPassword: newPassword,
    );
  }

  @override
  Future<void> refreshAccessToken(String refreshToken) {
    return remoteDataSource.refreshAccessToken(refreshToken);
  }

  @override
  Future<AuthResult> getMe() async {
    final result = await remoteDataSource.getMe();

    await localDataSource.saveQRCode(result.qrCode);
    await localDataSource.saveQRExpiresAt(result.user.qrExpiresAt!);

    return AuthResult(user: result.user, qrCode: result.qrCode);
  }
}

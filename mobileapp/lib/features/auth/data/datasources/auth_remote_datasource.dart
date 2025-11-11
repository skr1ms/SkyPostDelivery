import '../../../../core/network/http_client.dart';
import '../../../../core/network/api_constants.dart';
import '../../../../core/utils/phone_utils.dart';
import '../models/auth_result_model.dart';
import '../models/user_model.dart';

abstract class AuthRemoteDataSource {
  Future<String> register({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  });

  Future<AuthResultModel> login({
    required String login,
    required String password,
  });

  Future<String> loginByPhone({required String phone});

  Future<AuthResultModel> verifyPhone({
    required String phone,
    required String code,
  });

  Future<String> requestPasswordReset({required String phone});

  Future<String> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  });

  Future<void> refreshAccessToken(String refreshToken);

  Future<AuthResultModel> getMe();
}

class AuthRemoteDataSourceImpl implements AuthRemoteDataSource {
  final HttpClient httpClient;

  const AuthRemoteDataSourceImpl(this.httpClient);

  @override
  Future<String> register({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  }) async {
    final response = await httpClient.post(
      ApiConstants.authRegister,
      body: {
        'full_name': fullName,
        'email': email,
        'phone': PhoneUtils.cleanPhone(phone),
        'password': password,
      },
    );
    return response['message'] as String;
  }

  @override
  Future<AuthResultModel> login({
    required String login,
    required String password,
  }) async {
    final response = await httpClient.post(
      ApiConstants.authLogin,
      body: {'email': login, 'password': password},
    );

    final authResult = AuthResultModel.fromJson(response);
    if (authResult.accessToken != null && authResult.refreshToken != null) {
      httpClient.setTokens(authResult.accessToken!, authResult.refreshToken!);
    }
    return authResult;
  }

  @override
  Future<String> loginByPhone({required String phone}) async {
    final response = await httpClient.post(
      ApiConstants.authLoginPhone,
      body: {'phone': PhoneUtils.cleanPhone(phone)},
    );
    return response['message'] as String;
  }

  @override
  Future<AuthResultModel> verifyPhone({
    required String phone,
    required String code,
  }) async {
    final response = await httpClient.post(
      ApiConstants.authVerifyPhone,
      body: {'phone': PhoneUtils.cleanPhone(phone), 'code': code},
    );

    final authResult = AuthResultModel.fromJson(response);
    if (authResult.accessToken != null && authResult.refreshToken != null) {
      httpClient.setTokens(authResult.accessToken!, authResult.refreshToken!);
    }
    return authResult;
  }

  @override
  Future<String> requestPasswordReset({required String phone}) async {
    final response = await httpClient.post(
      ApiConstants.authPasswordResetRequest,
      body: {'phone': PhoneUtils.cleanPhone(phone)},
    );
    return response['message'] as String;
  }

  @override
  Future<String> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  }) async {
    final response = await httpClient.post(
      ApiConstants.authPasswordReset,
      body: {
        'phone': PhoneUtils.cleanPhone(phone),
        'code': code,
        'new_password': newPassword,
      },
    );
    return response['message'] as String;
  }

  @override
  Future<void> refreshAccessToken(String refreshToken) async {
    final response = await httpClient.post(ApiConstants.authRefresh, body: {});

    final accessToken = response['access_token'] as String;
    final newRefreshToken = response['refresh_token'] as String;
    httpClient.setTokens(accessToken, newRefreshToken);
  }

  @override
  Future<AuthResultModel> getMe() async {
    final response = await httpClient.get(
      ApiConstants.authMe,
      requiresAuth: true,
    );

    final user = UserModel.fromJson(response['user'] as Map<String, dynamic>);
    final qrCode = response['qr_code'] as String;

    return AuthResultModel(user: user, qrCode: qrCode);
  }
}

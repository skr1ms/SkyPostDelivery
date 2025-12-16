import '../../../../core/network/api_constants.dart';
import '../../../../core/network/http_client.dart';

abstract class NotificationRemoteDataSource {
  Future<void> registerDevice({
    required String userId,
    required String token,
    required String platform,
  });
}

class NotificationRemoteDataSourceImpl implements NotificationRemoteDataSource {
  final HttpClient httpClient;

  const NotificationRemoteDataSourceImpl(this.httpClient);

  @override
  Future<void> registerDevice({
    required String userId,
    required String token,
    required String platform,
  }) async {
    await httpClient.post(
      ApiConstants.userDevices(userId),
      body: {
        'token': token,
        'platform': platform,
      },
      requiresAuth: true,
    );
  }
}


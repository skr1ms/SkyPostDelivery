import '../../domain/repositories/notification_repository.dart';
import '../datasources/notification_remote_datasource.dart';

class NotificationRepositoryImpl implements NotificationRepository {
  const NotificationRepositoryImpl(this.remoteDataSource);

  final NotificationRemoteDataSource remoteDataSource;

  @override
  Future<void> registerDevice({
    required String userId,
    required String token,
    required String platform,
  }) {
    return remoteDataSource.registerDevice(
      userId: userId,
      token: token,
      platform: platform,
    );
  }
}


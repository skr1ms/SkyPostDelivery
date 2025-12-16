import '../repositories/notification_repository.dart';

class RegisterDeviceUseCase {
  const RegisterDeviceUseCase(this.repository);

  final NotificationRepository repository;

  Future<void> call({
    required String userId,
    required String token,
    required String platform,
  }) {
    return repository.registerDevice(
      userId: userId,
      token: token,
      platform: platform,
    );
  }
}


abstract class NotificationRepository {
  Future<void> registerDevice({
    required String userId,
    required String token,
    required String platform,
  });
}


import 'dart:async';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
}

class PushNotificationService {
  PushNotificationService(this._localNotifications);

  final FlutterLocalNotificationsPlugin _localNotifications;
  final AndroidNotificationChannel _androidChannel = const AndroidNotificationChannel(
    'delivery_updates',
    'Delivery Updates',
    description: 'Notifications about order status',
    importance: Importance.high,
  );
  bool _initialized = false;

  Future<void> initialize() async {
    if (_initialized) return;
    final messaging = FirebaseMessaging.instance;
    await messaging.requestPermission(alert: true, badge: true, sound: true);
    await messaging.setForegroundNotificationPresentationOptions(alert: true, badge: true, sound: true);
    const androidInit = AndroidInitializationSettings('@mipmap/ic_launcher');
    const darwinInit = DarwinInitializationSettings();
    const settings = InitializationSettings(android: androidInit, iOS: darwinInit);
    await _localNotifications.initialize(settings);
    final androidPlugin = _localNotifications.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
    if (androidPlugin != null) {
      await androidPlugin.createNotificationChannel(_androidChannel);
    }
    FirebaseMessaging.onMessage.listen(_onMessage);
    _initialized = true;
  }

  Future<void> _onMessage(RemoteMessage message) async {
    final notification = message.notification;
    if (notification == null) return;
    final android = notification.android;
    final details = NotificationDetails(
      android: AndroidNotificationDetails(
        _androidChannel.id,
        _androidChannel.name,
        channelDescription: _androidChannel.description,
        importance: Importance.high,
        priority: Priority.high,
        icon: android?.smallIcon,
      ),
      iOS: const DarwinNotificationDetails(),
    );
    await _localNotifications.show(
      notification.hashCode,
      notification.title,
      notification.body,
      details,
      payload: message.data['order_id'],
    );
  }

  Future<String?> getToken() {
    return FirebaseMessaging.instance.getToken();
  }

  Stream<String> get tokenRefreshStream {
    return FirebaseMessaging.instance.onTokenRefresh;
  }
}


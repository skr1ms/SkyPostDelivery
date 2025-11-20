import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

import 'core/theme/app_theme.dart';
import 'core/di/injection_container.dart';
import 'core/services/push_notification_service.dart';
import 'features/auth/presentation/providers/auth_provider.dart';
import 'features/auth/presentation/screens/splash_screen.dart';
import 'features/auth/presentation/screens/auth_screen.dart';
import 'features/home/presentation/screens/main_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);

  const environment = String.fromEnvironment(
    'ENVIRONMENT',
    defaultValue: 'local',
  );
  final envFile = environment == 'production'
      ? '.env.production'
      : '.env.local';

  try {
    await dotenv.load(fileName: envFile);
  } catch (e) {
    rethrow;
  }

  await Firebase.initializeApp();

  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.light,
      systemNavigationBarColor: AppTheme.backgroundColor,
      systemNavigationBarIconBrightness: Brightness.light,
    ),
  );

  SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  final di = InjectionContainer();
  await di.init();

  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => AuthProvider(
            loginUseCase: di.loginUseCase,
            registerUseCase: di.registerUseCase,
            verifyPhoneUseCase: di.verifyPhoneUseCase,
            getMeUseCase: di.getMeUseCase,
            getMyQRUseCase: di.getMyQRUseCase,
            refreshQRUseCase: di.refreshQRUseCase,
            localDataSource: di.authLocalDataSource,
            httpClient: di.httpClient,
            connectivityService: di.connectivityService,
            registerDeviceUseCase: di.registerDeviceUseCase,
            pushNotificationService: di.pushNotificationService,
          ),
        ),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'SkyPost Delivery',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme,
      home: const SplashScreen(),
      routes: {
        '/auth': (context) => const AuthScreen(),
        '/main': (context) => const MainScreen(),
      },
    );
  }
}

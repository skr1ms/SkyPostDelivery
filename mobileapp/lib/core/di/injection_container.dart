import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

import '../network/http_client.dart';
import '../services/connectivity_service.dart';
import '../services/push_notification_service.dart';

import '../../features/auth/data/datasources/auth_remote_datasource.dart';
import '../../features/auth/data/datasources/auth_local_datasource.dart';
import '../../features/auth/data/repositories/auth_repository_impl.dart';
import '../../features/auth/domain/repositories/auth_repository.dart';
import '../../features/auth/domain/usecases/login_usecase.dart';
import '../../features/auth/domain/usecases/register_usecase.dart';
import '../../features/auth/domain/usecases/verify_phone_usecase.dart';
import '../../features/auth/domain/usecases/get_me_usecase.dart';

import '../../features/goods/data/datasources/goods_remote_datasource.dart';
import '../../features/goods/data/repositories/goods_repository_impl.dart';
import '../../features/goods/domain/repositories/goods_repository.dart';
import '../../features/goods/domain/usecases/get_goods_usecase.dart';

import '../../features/orders/data/datasources/orders_remote_datasource.dart';
import '../../features/orders/data/repositories/orders_repository_impl.dart';
import '../../features/orders/domain/repositories/orders_repository.dart';
import '../../features/orders/domain/usecases/create_order_usecase.dart';
import '../../features/orders/domain/usecases/create_multiple_orders_usecase.dart';
import '../../features/orders/domain/usecases/get_user_orders_usecase.dart';
import '../../features/orders/domain/usecases/return_order_usecase.dart';

import '../../features/qr/data/datasources/qr_remote_datasource.dart';
import '../../features/qr/data/repositories/qr_repository_impl.dart';
import '../../features/qr/domain/repositories/qr_repository.dart';
import '../../features/qr/domain/usecases/get_my_qr_usecase.dart';
import '../../features/qr/domain/usecases/refresh_qr_usecase.dart';
import '../../features/notifications/data/datasources/notification_remote_datasource.dart';
import '../../features/notifications/data/repositories/notification_repository_impl.dart';
import '../../features/notifications/domain/repositories/notification_repository.dart';
import '../../features/notifications/domain/usecases/register_device_usecase.dart';

class InjectionContainer {
  static final InjectionContainer _instance = InjectionContainer._internal();
  factory InjectionContainer() => _instance;
  InjectionContainer._internal();

  late final HttpClient httpClient;
  late final FlutterSecureStorage storage;
  late final ConnectivityService connectivityService;
  late final FlutterLocalNotificationsPlugin localNotificationsPlugin;
  late final PushNotificationService pushNotificationService;

  late final AuthRepository authRepository;
  late final GoodsRepository goodsRepository;
  late final OrdersRepository ordersRepository;
  late final QRRepository qrRepository;
  late final NotificationRepository notificationRepository;

  late final LoginUseCase loginUseCase;
  late final RegisterUseCase registerUseCase;
  late final VerifyPhoneUseCase verifyPhoneUseCase;
  late final GetMeUseCase getMeUseCase;
  late final GetGoodsUseCase getGoodsUseCase;
  late final CreateOrderUseCase createOrderUseCase;
  late final CreateMultipleOrdersUseCase createMultipleOrdersUseCase;
  late final GetUserOrdersUseCase getUserOrdersUseCase;
  late final ReturnOrderUseCase returnOrderUseCase;
  late final GetMyQRUseCase getMyQRUseCase;
  late final RefreshQRUseCase refreshQRUseCase;
  late final RegisterDeviceUseCase registerDeviceUseCase;

  late final AuthLocalDataSource authLocalDataSource;

  Future<void> init() async {
    httpClient = HttpClient();
    storage = const FlutterSecureStorage();
    connectivityService = ConnectivityService();
    localNotificationsPlugin = FlutterLocalNotificationsPlugin();
    pushNotificationService = PushNotificationService(localNotificationsPlugin);
    await pushNotificationService.initialize();

    final authRemoteDataSource = AuthRemoteDataSourceImpl(httpClient);
    authLocalDataSource = AuthLocalDataSourceImpl(storage);
    authRepository = AuthRepositoryImpl(
      remoteDataSource: authRemoteDataSource,
      localDataSource: authLocalDataSource,
    );

    final goodsRemoteDataSource = GoodsRemoteDataSourceImpl(httpClient);
    goodsRepository = GoodsRepositoryImpl(goodsRemoteDataSource);

    final ordersRemoteDataSource = OrdersRemoteDataSourceImpl(httpClient);
    ordersRepository = OrdersRepositoryImpl(ordersRemoteDataSource);

    final qrRemoteDataSource = QRRemoteDataSourceImpl(httpClient);
    qrRepository = QRRepositoryImpl(qrRemoteDataSource);

    final notificationRemoteDataSource = NotificationRemoteDataSourceImpl(httpClient);
    notificationRepository = NotificationRepositoryImpl(notificationRemoteDataSource);

    loginUseCase = LoginUseCase(authRepository);
    registerUseCase = RegisterUseCase(authRepository);
    verifyPhoneUseCase = VerifyPhoneUseCase(authRepository);
    getMeUseCase = GetMeUseCase(authRepository);
    
    getGoodsUseCase = GetGoodsUseCase(goodsRepository);
    
    createOrderUseCase = CreateOrderUseCase(ordersRepository);
    createMultipleOrdersUseCase = CreateMultipleOrdersUseCase(ordersRepository);
    getUserOrdersUseCase = GetUserOrdersUseCase(ordersRepository);
    returnOrderUseCase = ReturnOrderUseCase(ordersRepository);
    
    getMyQRUseCase = GetMyQRUseCase(qrRepository);
    refreshQRUseCase = RefreshQRUseCase(qrRepository);
    registerDeviceUseCase = RegisterDeviceUseCase(notificationRepository);
  }
}


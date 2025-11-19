import 'dart:async';
import 'package:flutter/foundation.dart';
import '../../../../core/network/http_client.dart';
import '../../../../core/services/connectivity_service.dart';
import '../../../../core/services/push_notification_service.dart';
import '../../domain/entities/user_entity.dart';
import '../../domain/usecases/login_usecase.dart';
import '../../domain/usecases/register_usecase.dart';
import '../../domain/usecases/verify_phone_usecase.dart';
import '../../domain/usecases/get_me_usecase.dart';
import '../../data/datasources/auth_local_datasource.dart';
import '../../../notifications/domain/usecases/register_device_usecase.dart';

class AuthProvider with ChangeNotifier {
  final LoginUseCase loginUseCase;
  final RegisterUseCase registerUseCase;
  final VerifyPhoneUseCase verifyPhoneUseCase;
  final GetMeUseCase getMeUseCase;
  final AuthLocalDataSource localDataSource;
  final HttpClient httpClient;
  final ConnectivityService connectivityService;
  final RegisterDeviceUseCase registerDeviceUseCase;
  final PushNotificationService pushNotificationService;

  UserEntity? _user;
  String? _qrCode;
  bool _isAuthenticated = false;
  bool _isLoading = false;
  bool _isOfflineMode = false;
  StreamSubscription<String>? _tokenSubscription;

  AuthProvider({
    required this.loginUseCase,
    required this.registerUseCase,
    required this.verifyPhoneUseCase,
    required this.getMeUseCase,
    required this.localDataSource,
    required this.httpClient,
    required this.connectivityService,
    required this.registerDeviceUseCase,
    required this.pushNotificationService,
  }) {
    httpClient.setOnTokensRefreshedCallback(_onTokensRefreshed);
    _initConnectivityListener();
    _initNotificationListener();
  }

  void _initConnectivityListener() {
    connectivityService.connectionStream.listen((isConnected) {
      final wasOffline = _isOfflineMode;
      _isOfflineMode = !isConnected;
      notifyListeners();

      if (wasOffline && isConnected && _isAuthenticated) {
        _syncWithServer();
      }
    });
  }

  void _initNotificationListener() {
    _tokenSubscription = pushNotificationService.tokenRefreshStream.listen((
      token,
    ) {
      if (_user != null) {
        _registerDeviceToken(token);
      }
    });
  }

  Future<void> _syncWithServer() async {
    try {
      final result = await getMeUseCase();
      _user = result.user;
      _qrCode = result.qrCode;
      await localDataSource.saveUser(result.user);
      await localDataSource.saveQRCode(result.qrCode);
      notifyListeners();
      await _syncDeviceToken();
    } catch (e) {
      rethrow;
    }
  }

  void _onTokensRefreshed(String accessToken, String refreshToken) async {
    await localDataSource.saveTokens(
      accessToken: accessToken,
      refreshToken: refreshToken,
    );
    await _syncDeviceToken();
  }

  UserEntity? get user => _user;
  String? get qrCode => _qrCode;
  bool get isAuthenticated => _isAuthenticated;
  bool get isLoading => _isLoading;
  bool get isOfflineMode => _isOfflineMode;

  Future<void> login(String login, String password) async {
    _isLoading = true;
    notifyListeners();

    try {
      final result = await loginUseCase(login: login, password: password);
      _user = result.user;
      _qrCode = result.qrCode;
      _isAuthenticated = true;

      await localDataSource.saveUser(result.user);

      await _syncDeviceToken();
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      rethrow;
    }
  }

  Future<void> register({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  }) async {
    _isLoading = true;
    notifyListeners();

    try {
      await registerUseCase(
        fullName: fullName,
        email: email,
        phone: phone,
        password: password,
      );
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      rethrow;
    }
  }

  Future<void> verifyPhone(String phone, String code) async {
    _isLoading = true;
    notifyListeners();

    try {
      final result = await verifyPhoneUseCase(phone: phone, code: code);
      _user = result.user;
      _qrCode = result.qrCode;
      _isAuthenticated = true;

      await localDataSource.saveUser(result.user);

      await _syncDeviceToken();
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      rethrow;
    }
  }

  Future<void> logout() async {
    _user = null;
    _qrCode = null;
    _isAuthenticated = false;

    await localDataSource.clearAll();
    httpClient.clearTokens();

    notifyListeners();
  }

  Future<void> loadStoredAuth() async {
    final accessToken = await localDataSource.getAccessToken();
    final refreshToken = await localDataSource.getRefreshToken();
    final storedQrCode = await localDataSource.getQRCode();
    final storedUser = await localDataSource.getUser();

    if (accessToken != null && refreshToken != null) {
      httpClient.setTokens(accessToken, refreshToken);

      final hasConnection = await connectivityService.checkConnection();

      if (hasConnection) {
        try {
          final result = await getMeUseCase();
          _user = result.user;
          _qrCode = result.qrCode;
          _isAuthenticated = true;
          _isOfflineMode = false;

          await localDataSource.saveUser(_user!);
          await localDataSource.saveQRCode(result.qrCode);
          notifyListeners();
          await _syncDeviceToken();
          return;
        } catch (e) {
          if (storedUser != null && storedQrCode != null) {
            _user = storedUser;
            _qrCode = storedQrCode;
            _isAuthenticated = true;
            _isOfflineMode = true;
            notifyListeners();
            await _syncDeviceToken();
            return;
          }
        }
      }

      if (storedUser != null && storedQrCode != null) {
        _user = storedUser;
        _qrCode = storedQrCode;
        _isAuthenticated = true;
        _isOfflineMode = true;
        notifyListeners();
        await _syncDeviceToken();
      } else {
        await logout();
      }
    }
  }

  bool get isQRExpired {
    if (_user?.qrExpiresAt == null) return true;
    return DateTime.now().isAfter(_user!.qrExpiresAt!);
  }

  Duration? get qrTimeRemaining {
    if (_user?.qrExpiresAt == null) return null;
    final remaining = _user!.qrExpiresAt!.difference(DateTime.now());
    return remaining.isNegative ? Duration.zero : remaining;
  }

  void updateQRCode(String newQRCode, {DateTime? expiresAt}) async {
    _qrCode = newQRCode;
    if (expiresAt != null && _user != null) {
      _user = UserEntity(
        id: _user!.id,
        fullName: _user!.fullName,
        email: _user!.email,
        phoneNumber: _user!.phoneNumber,
        phoneVerified: _user!.phoneVerified,
        role: _user!.role,
        createdAt: _user!.createdAt,
        qrIssuedAt: DateTime.now(),
        qrExpiresAt: expiresAt,
      );

      await localDataSource.saveUser(_user!);
      await localDataSource.saveQRExpiresAt(expiresAt);
    }
    notifyListeners();
  }

  Future<void> _syncDeviceToken() async {
    if (_user == null) return;
    final token = await pushNotificationService.getToken();
    if (token != null) {
      await _registerDeviceToken(token);
    }
  }

  Future<void> _registerDeviceToken(String token) async {
    if (_user == null) return;
    final storedToken = await localDataSource.getDeviceToken();
    if (storedToken == token) return;
    final platform = _resolvePlatform();
    try {
      await registerDeviceUseCase(
        userId: _user!.id,
        token: token,
        platform: platform,
      );
      await localDataSource.saveDeviceToken(token);
    } catch (_) {}
  }

  String _resolvePlatform() {
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return 'android';
      case TargetPlatform.iOS:
        return 'ios';
      case TargetPlatform.macOS:
        return 'macos';
      case TargetPlatform.windows:
        return 'windows';
      case TargetPlatform.linux:
        return 'linux';
      case TargetPlatform.fuchsia:
        return 'fuchsia';
    }
  }

  @override
  void dispose() {
    _tokenSubscription?.cancel();
    super.dispose();
  }
}

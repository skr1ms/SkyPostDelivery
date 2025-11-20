class ApiConstants {
  static const String authRegister = '/auth/register';
  static const String authLogin = '/auth/login';
  static const String authLoginPhone = '/auth/login/phone';
  static const String authVerifyPhone = '/auth/verify/phone';
  static const String authRefresh = '/auth/refresh';
  static const String authMe = '/auth/me';
  static const String authPasswordResetRequest = '/auth/password/reset/request';
  static const String authPasswordReset = '/auth/password/reset';
  
  static const String goods = '/goods';
  
  static const String orders = '/orders';
  static const String ordersBatch = '/orders/batch';
  static String orderById(String id) => '/orders/$id';
  static String ordersByUser(String userId) => '/orders/user/$userId';
  static String returnOrder(String orderId) => '/orders/$orderId/return';
  static String userDevices(String userId) => '/users/$userId/devices';
  
  static const String qrMe = '/qr/me';
  static const String qrRefresh = '/qr/refresh';
}


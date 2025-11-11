import 'dart:io';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class ApiConfig {
  static String get baseUrl {
    const envUrl = String.fromEnvironment('API_BASE_URL');
    if (envUrl.isNotEmpty) {
      return envUrl;
    }

    final dotenvUrl = dotenv.maybeGet('API_BASE_URL');
    if (dotenvUrl != null && dotenvUrl.isNotEmpty) {
      return dotenvUrl;
    }

    if (Platform.isAndroid) {
      return 'http://10.0.2.2:8080/api/v1'; // your ip address in local network
    } else if (Platform.isIOS) {
      return 'http://localhost:8080/api/v1';
    }
    return 'http://localhost:8080/api/v1';
  }

  static const Duration connectionTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);

  static const int maxRetries = 3;

  static String getFullUrl(String endpoint) {
    return '$baseUrl$endpoint';
  }
}

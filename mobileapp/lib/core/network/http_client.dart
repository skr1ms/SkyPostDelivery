import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../errors/exceptions.dart';

class HttpClient {
  final String baseUrl;
  String? _accessToken;
  String? _refreshToken;
  Function(String, String)? _onTokensRefreshed;
  bool _isRefreshing = false;

  HttpClient({String? baseUrl}) : baseUrl = baseUrl ?? ApiConfig.baseUrl;

  void setTokens(String accessToken, String refreshToken) {
    _accessToken = accessToken;
    _refreshToken = refreshToken;
  }

  void clearTokens() {
    _accessToken = null;
    _refreshToken = null;
  }

  void setOnTokensRefreshedCallback(Function(String, String) callback) {
    _onTokensRefreshed = callback;
  }

  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshToken;

  Map<String, String> _getHeaders({bool requiresAuth = false}) {
    final headers = {'Content-Type': 'application/json'};
    if (requiresAuth && _accessToken != null) {
      headers['Authorization'] = 'Bearer $_accessToken';
    }
    return headers;
  }

  Future<Map<String, dynamic>> get(
    String endpoint, {
    bool requiresAuth = false,
  }) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl$endpoint'),
        headers: _getHeaders(requiresAuth: requiresAuth),
      );
      return await _handleResponse(response, () async {
        return await http.get(
          Uri.parse('$baseUrl$endpoint'),
          headers: _getHeaders(requiresAuth: requiresAuth),
        );
      });
    } on ServerException {
      rethrow;
    } on AuthException {
      rethrow;
    } catch (e) {
      throw NetworkException('Network error: $e');
    }
  }

  Future<Map<String, dynamic>> post(
    String endpoint, {
    required Map<String, dynamic> body,
    bool requiresAuth = false,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl$endpoint'),
        headers: _getHeaders(requiresAuth: requiresAuth),
        body: jsonEncode(body),
      );
      return await _handleResponse(response, () async {
        return await http.post(
          Uri.parse('$baseUrl$endpoint'),
          headers: _getHeaders(requiresAuth: requiresAuth),
          body: jsonEncode(body),
        );
      });
    } on ServerException {
      rethrow;
    } on AuthException {
      rethrow;
    } catch (e) {
      throw NetworkException('Network error: $e');
    }
  }

  Future<List<dynamic>> getList(
    String endpoint, {
    bool requiresAuth = false,
  }) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl$endpoint'),
        headers: _getHeaders(requiresAuth: requiresAuth),
      );

      if (response.statusCode >= 200 && response.statusCode < 300) {
        return jsonDecode(response.body) as List<dynamic>;
      } else if (response.statusCode == 401) {
        if (_refreshToken != null && !_isRefreshing) {
          try {
            await _refreshTokens();
            final newResponse = await http.get(
              Uri.parse('$baseUrl$endpoint'),
              headers: _getHeaders(requiresAuth: requiresAuth),
            );
            if (newResponse.statusCode >= 200 && newResponse.statusCode < 300) {
              return jsonDecode(newResponse.body) as List<dynamic>;
            }
          } catch (e) {
            clearTokens();
            throw AuthException('Session expired. Please login again.');
          }
        }
        clearTokens();
        throw AuthException('Unauthorized');
      } else {
        final json = jsonDecode(response.body) as Map<String, dynamic>;
        throw ServerException(json['error'] as String? ?? 'Server error');
      }
    } on ServerException {
      rethrow;
    } on AuthException {
      rethrow;
    } catch (e) {
      throw NetworkException('Network error: $e');
    }
  }

  Future<List<dynamic>> postList(
    String endpoint, {
    required Map<String, dynamic> body,
    bool requiresAuth = false,
  }) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl$endpoint'),
        headers: _getHeaders(requiresAuth: requiresAuth),
        body: jsonEncode(body),
      );

      if (response.statusCode >= 200 && response.statusCode < 300) {
        return jsonDecode(response.body) as List<dynamic>;
      } else if (response.statusCode == 401) {
        if (_refreshToken != null && !_isRefreshing) {
          try {
            await _refreshTokens();
            final newResponse = await http.post(
              Uri.parse('$baseUrl$endpoint'),
              headers: _getHeaders(requiresAuth: requiresAuth),
              body: jsonEncode(body),
            );
            if (newResponse.statusCode >= 200 && newResponse.statusCode < 300) {
              return jsonDecode(newResponse.body) as List<dynamic>;
            }
          } catch (e) {
            clearTokens();
            throw AuthException('Session expired. Please login again.');
          }
        }
        clearTokens();
        throw AuthException('Unauthorized');
      } else {
        final json = jsonDecode(response.body) as Map<String, dynamic>;
        throw ServerException(json['error'] as String? ?? 'Server error');
      }
    } on ServerException {
      rethrow;
    } on AuthException {
      rethrow;
    } catch (e) {
      throw NetworkException('Network error: $e');
    }
  }

  Future<Map<String, dynamic>> _handleResponse(
    http.Response response,
    Future<http.Response> Function() retryRequest,
  ) async {
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    } else if (response.statusCode == 401) {
      if (_refreshToken != null && !_isRefreshing) {
        try {
          await _refreshTokens();
          final newResponse = await retryRequest();
          if (newResponse.statusCode >= 200 && newResponse.statusCode < 300) {
            return jsonDecode(newResponse.body) as Map<String, dynamic>;
          }
        } catch (e) {
          clearTokens();
          throw AuthException('Session expired. Please login again.');
        }
      }
      clearTokens();
      throw AuthException('Unauthorized');
    } else {
      final json = jsonDecode(response.body) as Map<String, dynamic>;
      throw ServerException(json['error'] as String? ?? 'Server error');
    }
  }

  Future<void> _refreshTokens() async {
    if (_isRefreshing) return;

    _isRefreshing = true;
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/auth/refresh'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_refreshToken',
        },
        body: jsonEncode({}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final newAccessToken = data['access_token'] as String;
        final newRefreshToken = data['refresh_token'] as String;

        setTokens(newAccessToken, newRefreshToken);

        if (_onTokensRefreshed != null) {
          _onTokensRefreshed!(newAccessToken, newRefreshToken);
        }
      } else {
        throw AuthException('Failed to refresh token');
      }
    } finally {
      _isRefreshing = false;
    }
  }
}

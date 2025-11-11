import '../../../../core/network/http_client.dart';
import '../../../../core/network/api_constants.dart';
import '../../domain/repositories/qr_repository.dart';

abstract class QRRemoteDataSource {
  Future<QRResult> refreshQR();
}

class QRRemoteDataSourceImpl implements QRRemoteDataSource {
  final HttpClient httpClient;

  const QRRemoteDataSourceImpl(this.httpClient);

  @override
  Future<QRResult> refreshQR() async {
    final response = await httpClient.post(
      ApiConstants.qrRefresh,
      body: {},
      requiresAuth: true,
    );
    
    return QRResult(
      qrCode: response['qr_code'] as String,
      expiresAt: response['expires_at'] as int,
    );
  }
}


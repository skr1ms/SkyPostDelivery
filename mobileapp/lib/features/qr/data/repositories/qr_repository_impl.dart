import '../../domain/repositories/qr_repository.dart';
import '../datasources/qr_remote_datasource.dart';

class QRRepositoryImpl implements QRRepository {
  final QRRemoteDataSource remoteDataSource;

  const QRRepositoryImpl(this.remoteDataSource);

  @override
  Future<QRResult> getMyQR() {
    return remoteDataSource.getMyQR();
  }

  @override
  Future<QRResult> refreshQR() {
    return remoteDataSource.refreshQR();
  }
}

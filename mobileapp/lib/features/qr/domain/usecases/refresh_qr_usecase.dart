import '../repositories/qr_repository.dart';

class RefreshQRUseCase {
  final QRRepository repository;

  const RefreshQRUseCase(this.repository);

  Future<QRResult> call() {
    return repository.refreshQR();
  }
}


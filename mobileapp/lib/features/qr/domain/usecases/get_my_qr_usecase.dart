import '../repositories/qr_repository.dart';

class GetMyQRUseCase {
  final QRRepository repository;

  const GetMyQRUseCase(this.repository);

  Future<QRResult> call() {
    return repository.getMyQR();
  }
}

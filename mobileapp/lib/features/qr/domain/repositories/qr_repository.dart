abstract class QRRepository {
  Future<QRResult> getMyQR();
  Future<QRResult> refreshQR();
}

class QRResult {
  final String qrCode;
  final int issuedAt;
  final int expiresAt;

  const QRResult({
    required this.qrCode,
    required this.issuedAt,
    required this.expiresAt,
  });
}

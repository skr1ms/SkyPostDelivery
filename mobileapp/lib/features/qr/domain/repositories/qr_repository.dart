abstract class QRRepository {
  Future<QRResult> refreshQR();
}

class QRResult {
  final String qrCode;
  final int expiresAt;

  const QRResult({
    required this.qrCode,
    required this.expiresAt,
  });
}


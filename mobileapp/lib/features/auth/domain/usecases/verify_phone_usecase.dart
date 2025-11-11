import '../entities/auth_result.dart';
import '../repositories/auth_repository.dart';

class VerifyPhoneUseCase {
  final AuthRepository repository;

  const VerifyPhoneUseCase(this.repository);

  Future<AuthResult> call({
    required String phone,
    required String code,
  }) {
    return repository.verifyPhone(phone: phone, code: code);
  }
}


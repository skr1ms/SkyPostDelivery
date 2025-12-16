import '../entities/auth_result.dart';
import '../repositories/auth_repository.dart';

class GetMeUseCase {
  final AuthRepository repository;

  const GetMeUseCase(this.repository);

  Future<AuthResult> call() {
    return repository.getMe();
  }
}

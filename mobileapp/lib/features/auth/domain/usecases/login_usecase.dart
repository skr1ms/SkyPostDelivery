import '../entities/auth_result.dart';
import '../repositories/auth_repository.dart';

class LoginUseCase {
  final AuthRepository repository;

  const LoginUseCase(this.repository);

  Future<AuthResult> call({
    required String login,
    required String password,
  }) {
    return repository.login(login: login, password: password);
  }
}


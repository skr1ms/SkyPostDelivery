import '../repositories/auth_repository.dart';

class RegisterUseCase {
  final AuthRepository repository;

  const RegisterUseCase(this.repository);

  Future<String> call({
    required String fullName,
    required String email,
    required String phone,
    required String password,
  }) {
    return repository.register(
      fullName: fullName,
      email: email,
      phone: phone,
      password: password,
    );
  }
}


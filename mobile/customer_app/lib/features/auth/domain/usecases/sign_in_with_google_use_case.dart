import '../models/auth_session.dart';
import '../repositories/auth_repository.dart';

class SignInWithGoogleUseCase {
  final AuthRepository repository;

  SignInWithGoogleUseCase(this.repository);

  Future<AuthSession> call() {
    return repository.signInWithGoogle();
  }
}

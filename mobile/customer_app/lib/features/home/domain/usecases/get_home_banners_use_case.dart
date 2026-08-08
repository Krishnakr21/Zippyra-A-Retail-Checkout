import '../entities/home_banner.dart';
import '../repositories/home_repository.dart';

class GetHomeBannersUseCase {
  final HomeRepository repository;

  GetHomeBannersUseCase(this.repository);

  Future<List<HomeBanner>> call() {
    return repository.getHomeBanners();
  }
}

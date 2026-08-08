import '../entities/home_banner.dart';

abstract class HomeRepository {
  Future<List<HomeBanner>> getHomeBanners();
}

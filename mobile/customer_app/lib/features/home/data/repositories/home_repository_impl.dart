import '../../domain/entities/home_banner.dart';
import '../../domain/repositories/home_repository.dart';
import '../datasources/home_remote_data_source.dart';

class HomeRepositoryImpl implements HomeRepository {
  final HomeRemoteDataSource remoteDataSource;

  HomeRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<HomeBanner>> getHomeBanners() async {
    try {
      return await remoteDataSource.getHomeBanners();
    } catch (e) {
      return [];
    }
  }
}

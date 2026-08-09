import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/home/domain/entities/home_banner.dart';
import 'package:customer_app/features/home/domain/usecases/get_home_banners_use_case.dart';
import 'package:customer_app/features/home/data/repositories/home_repository_impl.dart';
import 'package:customer_app/features/home/data/datasources/home_remote_data_source.dart';
import 'package:customer_app/features/home/data/models/home_banner_model.dart';

class MockHomeRemoteDataSourceEmpty implements HomeRemoteDataSource {
  @override
  Future<List<HomeBannerModel>> getHomeBanners() async {
    return [];
  }
}

void main() {
  test('GetHomeBannersUseCase handles zero active banners gracefully', () async {
    final ds = MockHomeRemoteDataSourceEmpty();
    final repo = HomeRepositoryImpl(remoteDataSource: ds);
    final useCase = GetHomeBannersUseCase(repo);

    final banners = await useCase();

    expect(banners, isEmpty);
  });
}

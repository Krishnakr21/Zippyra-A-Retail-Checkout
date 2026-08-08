import 'package:zippyra_core/zippyra_core.dart';
import '../models/home_banner_model.dart';

abstract class HomeRemoteDataSource {
  Future<List<HomeBannerModel>> getHomeBanners();
}

class HomeRemoteDataSourceImpl implements HomeRemoteDataSource {
  final ApiClient apiClient;

  HomeRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<HomeBannerModel>> getHomeBanners() async {
    try {
      final response = await apiClient.get('/v1/store/home-banners');
      final data = response.data as Map<String, dynamic>;
      final list = data['banners'] as List<dynamic>? ?? [];
      return list.map((json) => HomeBannerModel.fromJson(json as Map<String, dynamic>)).toList();
    } catch (e) {
      // Fallback empty list on error for clean graceful UI
      return [];
    }
  }
}

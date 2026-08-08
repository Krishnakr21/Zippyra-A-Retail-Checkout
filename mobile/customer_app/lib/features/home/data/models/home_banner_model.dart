import '../../domain/entities/home_banner.dart';

class HomeBannerModel extends HomeBanner {
  const HomeBannerModel({
    required super.id,
    required super.title,
    required super.imageUrl,
    required super.deepLink,
  });

  factory HomeBannerModel.fromJson(Map<String, dynamic> json) {
    return HomeBannerModel(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      imageUrl: json['image_url'] as String? ?? '',
      deepLink: json['deep_link'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'image_url': imageUrl,
      'deep_link': deepLink,
    };
  }
}

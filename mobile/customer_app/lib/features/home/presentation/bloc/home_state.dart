import 'package:equatable/equatable.dart';
import '../../domain/entities/home_banner.dart';
import '../../../store_session/domain/entities/nearby_store.dart';

abstract class HomeState extends Equatable {
  const HomeState();

  @override
  List<Object?> get props => [];
}

class HomeInitialState extends HomeState {}

class HomeLoadingState extends HomeState {}

class HomeLoadedState extends HomeState {
  final List<HomeBanner> banners;
  final List<NearbyStore> nearbyStores;
  final List<Map<String, String>> recentStores;

  const HomeLoadedState({
    required this.banners,
    required this.nearbyStores,
    required this.recentStores,
  });

  @override
  List<Object?> get props => [banners, nearbyStores, recentStores];
}

class HomeErrorState extends HomeState {
  final String message;

  const HomeErrorState({required this.message});

  @override
  List<Object?> get props => [message];
}

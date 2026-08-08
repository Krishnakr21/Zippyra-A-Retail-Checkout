import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/usecases/get_home_banners_use_case.dart';
import '../../../store_session/domain/usecases/get_nearby_stores_use_case.dart';
import '../../../store_session/domain/entities/nearby_store.dart';
import '../../domain/entities/home_banner.dart';
import 'home_event.dart';
import 'home_state.dart';

class HomeBloc extends Bloc<HomeEvent, HomeState> {
  final GetHomeBannersUseCase getHomeBannersUseCase;
  final GetNearbyStoresUseCase getNearbyStoresUseCase;

  HomeBloc({
    required this.getHomeBannersUseCase,
    required this.getNearbyStoresUseCase,
  }) : super(HomeInitialState()) {
    on<LoadHomeDataEvent>(_onLoadHomeData);
  }

  Future<void> _onLoadHomeData(
    LoadHomeDataEvent event,
    Emitter<HomeState> emit,
  ) async {
    emit(HomeLoadingState());

    List<HomeBanner> banners = [];
    try {
      banners = await getHomeBannersUseCase();
    } catch (_) {
      banners = [];
    }

    List<NearbyStore> nearbyStores = [];
    try {
      nearbyStores = await getNearbyStoresUseCase(event.lat, event.lng);
    } catch (_) {
      nearbyStores = [];
    }

    final recentStores = [
      {'name': 'Zippyra Indiranagar', 'city': 'Bengaluru', 'last_visited': 'Yesterday'},
      {'name': 'Zippyra Koramangala', 'city': 'Bengaluru', 'last_visited': '3 days ago'},
    ];

    emit(HomeLoadedState(
      banners: banners,
      nearbyStores: nearbyStores,
      recentStores: recentStores,
    ));
  }
}

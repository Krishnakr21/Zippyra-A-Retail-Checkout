import 'package:equatable/equatable.dart';

abstract class HomeEvent extends Equatable {
  const HomeEvent();

  @override
  List<Object?> get props => [];
}

class LoadHomeDataEvent extends HomeEvent {
  final double lat;
  final double lng;

  const LoadHomeDataEvent({this.lat = 12.9716, this.lng = 77.5946});

  @override
  List<Object?> get props => [lat, lng];
}

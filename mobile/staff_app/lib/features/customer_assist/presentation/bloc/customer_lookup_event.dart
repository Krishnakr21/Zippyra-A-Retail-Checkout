import 'package:equatable/equatable.dart';

abstract class CustomerLookupEvent extends Equatable {
  const CustomerLookupEvent();

  @override
  List<Object?> get props => [];
}

class LookupRequested extends CustomerLookupEvent {
  final String storeId;
  final String phoneLast4;

  const LookupRequested({required this.storeId, required this.phoneLast4});

  @override
  List<Object?> get props => [storeId, phoneLast4];
}

class CustomerLookupReset extends CustomerLookupEvent {}

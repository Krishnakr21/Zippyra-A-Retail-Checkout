part of 'loyalty_bloc.dart';

abstract class LoyaltyEvent extends Equatable {
  const LoyaltyEvent();

  @override
  List<Object?> get props => [];
}

class LoyaltyBalanceRequested extends LoyaltyEvent {
  final bool refresh;

  const LoyaltyBalanceRequested({this.refresh = false});

  @override
  List<Object?> get props => [refresh];
}

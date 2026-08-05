import 'package:equatable/equatable.dart';

class CustomerSummary extends Equatable {
  final String customerId;
  final String firstName;
  final String phoneMasked;
  final String storeId;
  final bool hasActiveSession;
  final String? sessionId;
  final String? activeOrderId;
  final String? activeOrderStatus;

  const CustomerSummary({
    required this.customerId,
    required this.firstName,
    required this.phoneMasked,
    required this.storeId,
    required this.hasActiveSession,
    this.sessionId,
    this.activeOrderId,
    this.activeOrderStatus,
  });

  factory CustomerSummary.fromJson(Map<String, dynamic> json) {
    return CustomerSummary(
      customerId: json['customer_id'] as String? ?? '',
      firstName: json['first_name'] as String? ?? 'Customer',
      phoneMasked: json['phone_masked'] as String? ?? '',
      storeId: json['store_id'] as String? ?? '',
      hasActiveSession: json['has_active_session'] as bool? ?? false,
      sessionId: json['session_id'] as String?,
      activeOrderId: json['active_order_id'] as String?,
      activeOrderStatus: json['active_order_status'] as String?,
    );
  }

  @override
  List<Object?> get props => [
        customerId,
        firstName,
        phoneMasked,
        storeId,
        hasActiveSession,
        sessionId,
        activeOrderId,
        activeOrderStatus,
      ];
}

abstract class CustomerLookupState extends Equatable {
  const CustomerLookupState();

  @override
  List<Object?> get props => [];
}

class CustomerLookupInitial extends CustomerLookupState {}

class CustomerLookupSearching extends CustomerLookupState {}

class SingleMatch extends CustomerLookupState {
  final CustomerSummary customer;

  const SingleMatch(this.customer);

  @override
  List<Object?> get props => [customer];
}

class MultipleMatches extends CustomerLookupState {
  final List<CustomerSummary> candidates;

  const MultipleMatches(this.candidates);

  @override
  List<Object?> get props => [candidates];
}

class NoMatch extends CustomerLookupState {
  final String phoneLast4;

  const NoMatch(this.phoneLast4);

  @override
  List<Object?> get props => [phoneLast4];
}

class CustomerLookupFailed extends CustomerLookupState {
  final String message;

  const CustomerLookupFailed(this.message);

  @override
  List<Object?> get props => [message];
}

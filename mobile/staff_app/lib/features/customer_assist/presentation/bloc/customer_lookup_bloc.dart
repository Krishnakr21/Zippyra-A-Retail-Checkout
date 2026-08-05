import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'customer_lookup_event.dart';
import 'customer_lookup_state.dart';

class CustomerLookupBloc extends Bloc<CustomerLookupEvent, CustomerLookupState> {
  final ApiClient apiClient;

  CustomerLookupBloc({required this.apiClient}) : super(CustomerLookupInitial()) {
    on<LookupRequested>(_onLookupRequested);
    on<CustomerLookupReset>((event, emit) => emit(CustomerLookupInitial()));
  }

  Future<void> _onLookupRequested(
    LookupRequested event,
    Emitter<CustomerLookupState> emit,
  ) async {
    final phoneLast4 = event.phoneLast4.trim();
    if (phoneLast4.length != 4) {
      emit(const CustomerLookupFailed('Please enter exactly the last 4 digits of phone number'));
      return;
    }

    emit(CustomerLookupSearching());

    try {
      final response = await apiClient.get('/v1/order/internal/lookup-by-phone-last4', queryParameters: {
        'store_id': event.storeId,
        'phone_last4': phoneLast4,
      });

      if (response.statusCode == 200 && response.data != null) {
        final data = response.data as Map<String, dynamic>;
        final matchType = data['match_type'] as String? ?? 'NONE';

        if (matchType == 'SINGLE' && data['customer'] != null) {
          final customer = CustomerSummary.fromJson(data['customer'] as Map<String, dynamic>);
          emit(SingleMatch(customer));
          return;
        }

        if (matchType == 'MULTIPLE' && data['candidates'] != null) {
          final list = (data['candidates'] as List)
              .map((c) => CustomerSummary.fromJson(c as Map<String, dynamic>))
              .toList();
          emit(MultipleMatches(list));
          return;
        }
      }

      emit(NoMatch(phoneLast4));
    } catch (e) {
      emit(CustomerLookupFailed('Lookup failed: ${e.toString()}'));
    }
  }
}

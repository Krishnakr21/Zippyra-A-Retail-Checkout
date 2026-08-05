import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:staff_app/features/customer_assist/presentation/bloc/customer_lookup_bloc.dart';
import 'package:staff_app/features/customer_assist/presentation/bloc/customer_lookup_event.dart';
import 'package:staff_app/features/customer_assist/presentation/bloc/customer_lookup_state.dart';

class FakeLookupApiClient extends Fake implements ApiClient {
  @override
  Future<Response> get(String path, {Map<String, dynamic>? queryParameters}) async {
    final phoneLast4 = queryParameters?['phone_last4'];

    if (phoneLast4 == '5555') {
      return Response(
        requestOptions: RequestOptions(path: path),
        statusCode: 200,
        data: {
          'match_type': 'MULTIPLE',
          'candidates': [
            {
              'customer_id': 'user-1111115555',
              'first_name': 'Rahul',
              'phone_masked': '+91XXXXXX5555',
              'store_id': 'store-1',
              'has_active_session': true,
              'session_id': 'sess-1',
            },
            {
              'customer_id': 'user-9999995555',
              'first_name': 'Rohan',
              'phone_masked': '+91XXXXXX5555',
              'store_id': 'store-1',
              'has_active_session': false,
              'session_id': 'sess-2',
            },
          ],
        },
      );
    }

    if (phoneLast4 == '1234') {
      return Response(
        requestOptions: RequestOptions(path: path),
        statusCode: 200,
        data: {
          'match_type': 'SINGLE',
          'customer': {
            'customer_id': 'user-9876541234',
            'first_name': 'Priya',
            'phone_masked': '+91XXXXXX1234',
            'store_id': 'store-1',
            'has_active_session': true,
            'session_id': 'sess-1234',
            'active_order_id': 'ord-1234',
            'active_order_status': 'PAYMENT_PENDING',
          },
        },
      );
    }

    return Response(
      requestOptions: RequestOptions(path: path),
      statusCode: 200,
      data: {
        'match_type': 'NONE',
      },
    );
  }
}

void main() {
  late FakeLookupApiClient fakeApiClient;
  late CustomerLookupBloc bloc;

  setUp(() {
    fakeApiClient = FakeLookupApiClient();
    bloc = CustomerLookupBloc(apiClient: fakeApiClient);
  });

  group('CustomerLookupBloc Tests', () {
    test('last-4 query matching 2 customers returns MultipleMatches with masked phone numbers', () async {
      bloc.add(const LookupRequested(storeId: 'store-1', phoneLast4: '5555'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          CustomerLookupSearching(),
          predicate<MultipleMatches>((s) {
            return s.candidates.length == 2 &&
                s.candidates.every((c) => c.phoneMasked == '+91XXXXXX5555');
          }),
        ]),
      );
    });

    test('last-4 query matching 1 customer returns SingleMatch with detailed status', () async {
      bloc.add(const LookupRequested(storeId: 'store-1', phoneLast4: '1234'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          CustomerLookupSearching(),
          predicate<SingleMatch>((s) {
            return s.customer.customerId == 'user-9876541234' &&
                s.customer.activeOrderStatus == 'PAYMENT_PENDING' &&
                s.customer.phoneMasked == '+91XXXXXX1234';
          }),
        ]),
      );
    });

    test('non-4-digit input fails validation', () async {
      bloc.add(const LookupRequested(storeId: 'store-1', phoneLast4: '12'));

      await expectLater(
        bloc.stream,
        emitsInOrder([
          predicate<CustomerLookupFailed>((s) => s.message.contains('exactly the last 4 digits')),
        ]),
      );
    });
  });
}

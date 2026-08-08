import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/order_detail.dart';
import '../../domain/usecases/get_order_detail_use_case.dart';
import '../../domain/usecases/request_return_use_case.dart';

part 'order_detail_event.dart';
part 'order_detail_state.dart';

class OrderDetailBloc extends Bloc<OrderDetailEvent, OrderDetailState> {
  final GetOrderDetailUseCase getOrderDetailUseCase;
  final RequestReturnUseCase requestReturnUseCase;

  OrderDetailBloc({
    required this.getOrderDetailUseCase,
    required this.requestReturnUseCase,
  }) : super(OrderDetailInitial()) {
    on<OrderDetailRequested>(_onOrderDetailRequested);
    on<ReturnRequested>(_onReturnRequested);
  }

  Future<void> _onOrderDetailRequested(
    OrderDetailRequested event,
    Emitter<OrderDetailState> emit,
  ) async {
    emit(OrderDetailLoading());
    try {
      final order = await getOrderDetailUseCase(event.orderId);
      emit(OrderDetailLoaded(order));
    } catch (e) {
      emit(OrderDetailError(e.toString()));
    }
  }

  Future<void> _onReturnRequested(
    ReturnRequested event,
    Emitter<OrderDetailState> emit,
  ) async {
    OrderDetail? currentOrder;
    if (state is OrderDetailLoaded) {
      currentOrder = (state as OrderDetailLoaded).order;
    }

    emit(ReturnSubmitting());
    try {
      await requestReturnUseCase(
        orderId: event.orderId,
        itemBarcodes: event.itemBarcodes,
        reason: event.reason,
      );
      emit(ReturnSubmitted(event.orderId));
    } catch (e) {
      String code = ErrorCodes.invalidRequest;
      if (e is Failure && e.code != null) {
        code = e.code!;
      } else {
        final msg = e.toString();
        if (msg.contains(ErrorCodes.returnWindowExpired)) {
          code = ErrorCodes.returnWindowExpired;
        } else if (msg.contains(ErrorCodes.itemNotReturnable)) {
          code = ErrorCodes.itemNotReturnable;
        } else if (msg.contains(ErrorCodes.returnQtyExceeded)) {
          code = ErrorCodes.returnQtyExceeded;
        }
      }
      emit(ReturnFailed(errorCode: code, message: e.toString(), currentOrder: currentOrder));
    }
  }
}

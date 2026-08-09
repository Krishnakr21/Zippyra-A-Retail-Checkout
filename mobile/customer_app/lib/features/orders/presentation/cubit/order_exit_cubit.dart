import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/exit_token.dart';
import '../../domain/usecases/get_exit_token_use_case.dart';

part 'order_exit_state.dart';

class OrderExitCubit extends Cubit<OrderExitState> {
  final GetExitTokenUseCase getExitTokenUseCase;

  OrderExitCubit({required this.getExitTokenUseCase}) : super(OrderExitInitial());

  Future<ExitToken?> fetchExitToken(String storeId) async {
    emit(OrderExitLoading());
    try {
      final token = await getExitTokenUseCase(storeId: storeId);
      emit(OrderExitLoaded(token));
      return token;
    } catch (e) {
      emit(OrderExitError(e.toString()));
      return null;
    }
  }
}

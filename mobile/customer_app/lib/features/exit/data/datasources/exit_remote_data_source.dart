import 'package:zippyra_core/zippyra_core.dart';
import '../models/exit_status_model.dart';

abstract class ExitRemoteDataSource {
  Future<ExitStatusModel> getExitStatus(String orderId);
}

class ExitRemoteDataSourceImpl implements ExitRemoteDataSource {
  final ApiClient apiClient;

  ExitRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<ExitStatusModel> getExitStatus(String orderId) async {
    final response = await apiClient.get('/v1/exit/status/$orderId');
    final data = response.data as Map<String, dynamic>;
    return ExitStatusModel.fromJson(data);
  }
}

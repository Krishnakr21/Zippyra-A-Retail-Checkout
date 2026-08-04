import 'package:equatable/equatable.dart';

class StaffSession extends Equatable {
  final String token;
  final String staffId;
  final String role; // CASHIER | STOCK_ASSOCIATE | SECURITY | MANAGER
  final String storeId;
  final String storeName;
  final bool hasPinSet;

  const StaffSession({
    required this.token,
    required this.staffId,
    required this.role,
    required this.storeId,
    required this.storeName,
    this.hasPinSet = false,
  });

  @override
  List<Object?> get props => [token, staffId, role, storeId, storeName, hasPinSet];
}

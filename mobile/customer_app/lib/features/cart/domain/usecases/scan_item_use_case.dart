import '../entities/cart_summary.dart';
import '../repositories/cart_repository.dart';

class ScanItemUseCase {
  final CartRepository repository;

  ScanItemUseCase(this.repository);

  Future<CartSummary> call(String storeId, String barcode, {int qty = 1}) async {
    return await repository.scanItem(storeId, barcode, qty: qty);
  }
}

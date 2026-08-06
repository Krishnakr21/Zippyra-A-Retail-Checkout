import '../entities/product.dart';
import '../repositories/catalog_repository.dart';

class GetProductByBarcodeUseCase {
  final CatalogRepository repository;

  const GetProductByBarcodeUseCase(this.repository);

  Future<Product?> call(String storeId, String barcode) {
    return repository.getProductByBarcode(storeId, barcode);
  }
}

import 'package:dio/dio.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/category.dart';
import '../../domain/entities/product.dart';

abstract class CatalogRemoteDataSource implements CatalogRemoteSyncDataSource {
  Future<Product> getProductByBarcode(String storeId, String barcode);
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1});
  Future<List<Category>> getCategories(String chainId);
  @override
  Future<Map<String, dynamic>> postDeltaSync(String storeId, int sinceSeq);
}

class CatalogRemoteDataSourceImpl implements CatalogRemoteDataSource {
  final ApiClient apiClient;

  const CatalogRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<Product> getProductByBarcode(String storeId, String barcode) async {
    try {
      final response = await apiClient.get('/v1/catalog/barcode/$barcode', queryParameters: {'store_id': storeId});
      final data = response.data as Map<String, dynamic>;
      return Product(
        id: data['id'] as String,
        barcode: data['barcode'] as String,
        name: data['name'] as String,
        description: data['description'] as String? ?? '',
        pricePaise: (data['price_paise'] as num).toInt(),
        mrpPaise: (data['mrp_paise'] as num).toInt(),
        hsnCode: data['hsn_code'] as String? ?? '',
        gstRatePercent: (data['gst_rate_percent'] as num).toDouble(),
        imageUrl: data['image_url'] as String? ?? '',
        thumbnailUrl: data['thumbnail_url'] as String? ?? '',
        isReturnable: data['is_returnable'] as bool? ?? true,
      );
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<List<Product>> searchProducts(String storeId, String query, {String? categoryId, int page = 1}) async {
    try {
      final response = await apiClient.get('/v1/catalog/search', queryParameters: {
        'store_id': storeId,
        'q': query,
        if (categoryId != null) 'category_id': categoryId,
        'page': page,
      });
      final data = response.data as Map<String, dynamic>;
      final productsJson = (data['products'] as List? ?? []);
      return productsJson.map((json) {
        return Product(
          id: json['id'] as String,
          barcode: json['barcode'] as String,
          name: json['name'] as String,
          description: json['description'] as String? ?? '',
          categoryId: json['category_id'] as String?,
          pricePaise: (json['price_paise'] as num).toInt(),
          mrpPaise: (json['mrp_paise'] as num).toInt(),
          hsnCode: json['hsn_code'] as String? ?? '',
          gstRatePercent: (json['gst_rate_percent'] as num).toDouble(),
          imageUrl: json['image_url'] as String? ?? '',
          thumbnailUrl: json['thumbnail_url'] as String? ?? '',
          isReturnable: json['is_returnable'] as bool? ?? true,
        );
      }).toList();
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<List<Category>> getCategories(String chainId) async {
    try {
      final response = await apiClient.get('/v1/catalog/categories', queryParameters: {'chain_id': chainId});
      final data = response.data as List;
      return data.map((json) {
        return Category(
          id: json['id'] as String,
          name: json['name'] as String,
          parentId: json['parent_id'] as String?,
          sortOrder: (json['sort_order'] as num?)?.toInt() ?? 0,
        );
      }).toList();
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  @override
  Future<Map<String, dynamic>> postDeltaSync(String storeId, int sinceSeq) async {
    try {
      final response = await apiClient.post('/v1/catalog/sync', data: {
        'store_id': storeId,
        'since_seq': sinceSeq,
      });
      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Failure _handleDioError(DioException e) {
    if (e.response != null && e.response?.data is Map<String, dynamic>) {
      final errObj = e.response?.data['error'];
      if (errObj != null && errObj is Map<String, dynamic>) {
        final code = errObj['code'] as String?;
        final message = errObj['message'] as String? ?? 'An error occurred';

        switch (code) {
          case ErrorCodes.productNotFound:
            return ServerFailure(message, code: code);
          case ErrorCodes.barcodeInvalid:
            return ServerFailure(message, code: code);
          default:
            return ServerFailure(message, code: code);
        }
      }
    }
    return NetworkFailure('Network connection error: ${e.message}');
  }
}

class CatalogItem {
  final String barcode;
  final String name;
  final int pricePaise;
  final String? thumbnailUrl;

  const CatalogItem({
    required this.barcode,
    required this.name,
    required this.pricePaise,
    this.thumbnailUrl,
  });
}

class LocalCatalogCache {
  final Map<String, CatalogItem> _memoryCache = {};

  void cacheItem(CatalogItem item) {
    _memoryCache[item.barcode] = item;
  }

  CatalogItem? lookup(String barcode) {
    return _memoryCache[barcode];
  }
}

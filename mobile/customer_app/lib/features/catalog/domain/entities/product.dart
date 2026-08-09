class Product {
  final String id;
  final String barcode;
  final String name;
  final String description;
  final String? categoryId;
  final int pricePaise;
  final int mrpPaise;
  final String hsnCode;
  final double gstRatePercent;
  final String imageUrl;
  final String thumbnailUrl;
  final bool isReturnable;

  const Product({
    required this.id,
    required this.barcode,
    required this.name,
    this.description = '',
    this.categoryId,
    required this.pricePaise,
    required this.mrpPaise,
    this.hsnCode = '',
    this.gstRatePercent = 0.0,
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.isReturnable = true,
  });
}

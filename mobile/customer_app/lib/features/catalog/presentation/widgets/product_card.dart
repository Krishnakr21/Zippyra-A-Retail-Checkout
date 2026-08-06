import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/product.dart';

class ProductCard extends StatelessWidget {
  final Product product;
  final VoidCallback? onTap;
  final VoidCallback? onAddToCart;

  const ProductCard({
    super.key,
    required this.product,
    this.onTap,
    this.onAddToCart,
  });

  @override
  Widget build(BuildContext context) {
    final priceRs = (product.pricePaise / 100.0).round();
    final mrpRs = (product.mrpPaise / 100.0).round();
    final isDiscounted = product.mrpPaise > product.pricePaise;
    final discountPercent = isDiscounted
        ? (((product.mrpPaise - product.pricePaise) / product.mrpPaise) * 100).round()
        : 0;
    final savingsRs = isDiscounted ? ((product.mrpPaise - product.pricePaise) / 100.0).round() : 0;

    final imgUrl = product.thumbnailUrl.isNotEmpty ? product.thumbnailUrl : product.imageUrl;

    final String ratingStr = '4.${5 + (product.barcode.hashCode % 4)} (${1 + (product.barcode.hashCode % 15)}.${(product.barcode.hashCode % 9)}k)';

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(14),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(14),
          child: Stack(
            children: [
              Padding(
                padding: const EdgeInsets.all(10.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Product Image Frame
                    Expanded(
                      child: Container(
                        width: double.infinity,
                        padding: const EdgeInsets.all(6),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF9FAFB),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Center(
                          child: imgUrl.isNotEmpty
                              ? CachedNetworkImage(
                                  imageUrl: imgUrl,
                                  fit: BoxFit.contain,
                                  placeholder: (context, url) => const Center(
                                    child: SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(strokeWidth: 2, color: ZippyraColors.primaryBlue),
                                    ),
                                  ),
                                  errorWidget: (context, url, error) => const Icon(
                                    Icons.inventory_2_outlined,
                                    size: 40,
                                    color: Color(0xFF9CA3AF),
                                  ),
                                )
                              : const Icon(
                                  Icons.inventory_2_outlined,
                                  size: 40,
                                  color: Color(0xFF9CA3AF),
                                ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),

                    // Zepto Rating Tag
                    Row(
                      children: [
                        const Icon(Icons.star_rounded, color: Color(0xFF16A34A), size: 14),
                        const SizedBox(width: 2),
                        Text(
                          ratingStr,
                          style: const TextStyle(
                            fontSize: 10,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF4B5563),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),

                    // Product Name (Max 2 lines)
                    Text(
                      product.name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF1F2937),
                        height: 1.2,
                      ),
                    ),
                    const SizedBox(height: 2),

                    // Pack / Unit Line
                    const Text(
                      'Fresh Stock',
                      style: TextStyle(
                        fontSize: 10,
                        color: Color(0xFF6B7280),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 8),

                    // Pricing & Zepto ADD Button Row
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Text(
                                  '₹$priceRs',
                                  style: const TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w900,
                                    color: Color(0xFF111827),
                                  ),
                                ),
                                if (isDiscounted) ...[
                                  const SizedBox(width: 4),
                                  Text(
                                    '₹$mrpRs',
                                    style: const TextStyle(
                                      fontSize: 10,
                                      color: Color(0xFF9CA3AF),
                                      decoration: TextDecoration.lineThrough,
                                    ),
                                  ),
                                ],
                              ],
                            ),
                            if (savingsRs > 0) ...[
                              const SizedBox(height: 1),
                              Text(
                                '₹$savingsRs OFF',
                                style: const TextStyle(
                                  fontSize: 10,
                                  fontWeight: FontWeight.w800,
                                  color: Color(0xFF16A34A),
                                ),
                              ),
                            ],
                          ],
                        ),

                        // Zepto Outline ADD Button
                        GestureDetector(
                          onTap: onAddToCart ??
                              () {
                                context.push('/scan', extra: {
                                  'name': product.name,
                                  'barcode': product.barcode,
                                  'price': '₹$priceRs',
                                  'image_url': imgUrl,
                                });
                              },
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              border: Border.all(color: const Color(0xFFE02461), width: 1.5),
                              borderRadius: BorderRadius.circular(8),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(0xFFE02461).withOpacity(0.1),
                                  blurRadius: 4,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: const Text(
                              'ADD',
                              style: TextStyle(
                                color: Color(0xFFE02461),
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                letterSpacing: 0.5,
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),

              // Discount Tag Badge (Top Left)
              if (isDiscounted && discountPercent > 0)
                Positioned(
                  top: 8,
                  left: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                    decoration: BoxDecoration(
                      color: const Color(0xFFDCFCE7),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: const Color(0xFF86EFAC), width: 0.8),
                    ),
                    child: Text(
                      '$discountPercent% OFF',
                      style: const TextStyle(
                        color: Color(0xFF15803D),
                        fontSize: 9,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 0.2,
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

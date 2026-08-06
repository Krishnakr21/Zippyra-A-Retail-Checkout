import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../../../injection_container.dart';
import '../../../cart/presentation/bloc/cart_bloc.dart';
import '../../domain/entities/product.dart';

class ProductDetailScreen extends StatefulWidget {
  final Product product;
  final Function(Product)? onAddToCart;

  const ProductDetailScreen({
    super.key,
    required this.product,
    this.onAddToCart,
  });

  @override
  State<ProductDetailScreen> createState() => _ProductDetailScreenState();
}

class _ProductDetailScreenState extends State<ProductDetailScreen> {
  int _quantity = 1;

  @override
  Widget build(BuildContext context) {
    final pricePaise = widget.product.pricePaise > 0 ? widget.product.pricePaise : 28000;
    final mrpPaise = widget.product.mrpPaise > 0 ? widget.product.mrpPaise : 31000;
    final priceRs = (pricePaise / 100.0).toStringAsFixed(0);
    final mrpRs = (mrpPaise / 100.0).toStringAsFixed(0);
    final isDiscounted = mrpPaise > pricePaise;
    final savingsRs = (mrpPaise - pricePaise) ~/ 100;

    return Scaffold(
      backgroundColor: const Color(0xFFF4F5FA),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        title: Text(
          widget.product.name.isNotEmpty ? widget.product.name : 'Product Details',
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: ZippyraColors.textPrimary),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.share_outlined, color: ZippyraColors.textPrimary),
            onPressed: () {},
          ),
        ],
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Showcase Image Area with Tags
                    Container(
                      height: 200,
                      width: double.infinity,
                      decoration: const BoxDecoration(
                        gradient: LinearGradient(
                          colors: [Color(0xFFEBF4FF), Colors.white],
                          begin: Alignment.topCenter,
                          end: Alignment.bottomCenter,
                        ),
                      ),
                      child: Stack(
                        alignment: Alignment.center,
                        children: [
                          widget.product.imageUrl.isNotEmpty
                              ? CachedNetworkImage(
                                  imageUrl: widget.product.imageUrl,
                                  height: 160,
                                  fit: BoxFit.contain,
                                  placeholder: (context, url) => const Center(child: CircularProgressIndicator()),
                                  errorWidget: (context, url, error) => const Text('🧈', style: TextStyle(fontSize: 90)),
                                )
                              : const Text('🧈', style: TextStyle(fontSize: 90)),
                          Positioned(
                            top: 12,
                            left: 12,
                            child: Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: const Color(0xFFE8F1FB),
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: const Text('RFID TAGGED', style: TextStyle(color: ZippyraColors.primaryBlue, fontSize: 9, fontWeight: FontWeight.w900)),
                                ),
                                const SizedBox(width: 6),
                                if (isDiscounted)
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: ZippyraColors.accentOrange,
                                      borderRadius: BorderRadius.circular(6),
                                    ),
                                    child: const Text('10% OFF', style: TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.w900)),
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),

                    // Body Details
                    Container(
                      color: Colors.white,
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'AMUL',
                            style: TextStyle(fontSize: 10, fontWeight: FontWeight.w800, color: ZippyraColors.primaryBlue, letterSpacing: 1),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            widget.product.name.isNotEmpty ? widget.product.name : 'Amul Pasteurised Butter',
                            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w900, color: ZippyraColors.textPrimary),
                          ),
                          const SizedBox(height: 8),

                          Row(
                            crossAxisAlignment: CrossAxisAlignment.baseline,
                            textBaseline: TextBaseline.alphabetic,
                            children: [
                              Text(
                                '₹$priceRs',
                                style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w900, color: ZippyraColors.textPrimary),
                              ),
                              if (isDiscounted) ...[
                                const SizedBox(width: 8),
                                Text(
                                  '₹$mrpRs',
                                  style: const TextStyle(fontSize: 14, decoration: TextDecoration.lineThrough, color: ZippyraColors.textSecondary),
                                ),
                                const SizedBox(width: 8),
                                Text(
                                  'Save ₹$savingsRs',
                                  style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: ZippyraColors.successGreen),
                                ),
                              ],
                            ],
                          ),
                          const SizedBox(height: 16),

                          // Quantity Selector & Stock Status
                          Row(
                            children: [
                              Container(
                                decoration: BoxDecoration(
                                  border: Border.all(color: ZippyraColors.border),
                                  borderRadius: BorderRadius.circular(10),
                                ),
                                child: Row(
                                  children: [
                                    GestureDetector(
                                      onTap: () {
                                        if (_quantity > 1) setState(() => _quantity--);
                                      },
                                      child: Container(
                                        width: 32,
                                        height: 32,
                                        decoration: const BoxDecoration(
                                          color: Color(0xFFE8F1FB),
                                          borderRadius: BorderRadius.horizontal(left: Radius.circular(9)),
                                        ),
                                        child: const Center(child: Text('−', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: ZippyraColors.primaryBlue))),
                                      ),
                                    ),
                                    SizedBox(
                                      width: 36,
                                      child: Center(
                                        child: Text('$_quantity', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold)),
                                      ),
                                    ),
                                    GestureDetector(
                                      onTap: () => setState(() => _quantity++),
                                      child: Container(
                                        width: 32,
                                        height: 32,
                                        decoration: const BoxDecoration(
                                          color: Color(0xFFE8F1FB),
                                          borderRadius: BorderRadius.horizontal(right: Radius.circular(9)),
                                        ),
                                        child: const Center(child: Text('+', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: ZippyraColors.primaryBlue))),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 12),
                              const Text(
                                '✓ In Stock · Shelf B-12',
                                style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: ZippyraColors.successGreen),
                              ),
                            ],
                          ),
                          const SizedBox(height: 16),

                          // 3 Attribute Metric Badges
                          Row(
                            children: [
                              Expanded(child: _buildMetricTile(icon: '🕐', title: '12 mins', sub: 'Delivery')),
                              const SizedBox(width: 8),
                              Expanded(child: _buildMetricTile(icon: '⭐', title: '4.7', sub: 'Rating')),
                              const SizedBox(width: 8),
                              Expanded(child: _buildMetricTile(icon: '🌡️', title: 'Cold', sub: 'Stored')),
                            ],
                          ),
                          const SizedBox(height: 16),

                          Text(
                            widget.product.description.isNotEmpty
                                ? widget.product.description
                                : 'Amul Pasteurised Butter made from fresh cream. Rich in vitamins A, D, E. Best for cooking & baking. No artificial colors.',
                            style: const TextStyle(color: ZippyraColors.textSecondary, fontSize: 12, height: 1.5),
                          ),
                          const SizedBox(height: 14),

                          // HSN Code & GST Table Card
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF4F5FA),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: Column(
                              children: [
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text('HSN Code', style: TextStyle(fontSize: 10, color: ZippyraColors.textSecondary)),
                                    Text(widget.product.hsnCode.isNotEmpty ? widget.product.hsnCode : '0405', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: ZippyraColors.textPrimary)),
                                  ],
                                ),
                                const SizedBox(height: 4),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text('GST Rate', style: TextStyle(fontSize: 10, color: ZippyraColors.textSecondary)),
                                    Text('${widget.product.gstRatePercent > 0 ? widget.product.gstRatePercent : 12}%', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: ZippyraColors.textPrimary)),
                                  ],
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),

            // Bottom Sticky CTA Bar
            Container(
              padding: const EdgeInsets.all(14),
              decoration: const BoxDecoration(
                color: Colors.white,
                border: Border(top: BorderSide(color: ZippyraColors.border)),
              ),
              child: Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: const Color(0xFFE8F1FB),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.camera_alt_outlined, color: ZippyraColors.primaryBlue, size: 22),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: SizedBox(
                      height: 44,
                      child: ElevatedButton(
                        onPressed: () {
                          if (widget.onAddToCart != null) {
                            widget.onAddToCart!(widget.product);
                          } else {
                            sl<CartBloc>().add(ItemScanned(
                              storeId: 'store-1',
                              barcode: widget.product.barcode.isNotEmpty ? widget.product.barcode : '8901262010051',
                            ));
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text('${widget.product.name.isNotEmpty ? widget.product.name : 'Amul Butter'} added to cart!'),
                                behavior: SnackBarBehavior.floating,
                                backgroundColor: ZippyraColors.successGreen,
                              ),
                            );
                          }
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: ZippyraColors.primaryBlue,
                          foregroundColor: Colors.white,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          elevation: 0,
                        ),
                        child: Text(
                          'Add to Cart · ₹${(pricePaise * _quantity / 100.0).toStringAsFixed(0)}',
                          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMetricTile({required String icon, required String title, required String sub}) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: const Color(0xFFF4F5FA),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Column(
        children: [
          Text(icon, style: const TextStyle(fontSize: 16)),
          const SizedBox(height: 2),
          Text(title, style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: ZippyraColors.textPrimary)),
          Text(sub, style: const TextStyle(fontSize: 8, color: ZippyraColors.textSecondary)),
        ],
      ),
    );
  }
}


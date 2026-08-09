import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import '../../../cart/presentation/bloc/cart_bloc.dart';

class CategoryProductsScreen extends StatefulWidget {
  final String categoryName;

  const CategoryProductsScreen({super.key, required this.categoryName});

  @override
  State<CategoryProductsScreen> createState() => _CategoryProductsScreenState();
}

class _CategoryProductsScreenState extends State<CategoryProductsScreen> {
  static const List<Map<String, dynamic>> _allProducts = [
    {
      'name': '100% Rolled Oats (Yoga Bar)',
      'category': 'Breakfast & Sauces',
      'weight': '500g Pack',
      'price': '₹493',
      'mrp': '₹530',
      'savings': '₹37 OFF',
      'badge': '7% OFF',
      'rating': '4.6 (14.2k)',
      'barcode': '8904335601951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
    },
    {
      'name': 'Britannia Toastea Bake Rusk',
      'category': 'Packaged Food',
      'weight': '250g Pack',
      'price': '₹212',
      'mrp': '₹237',
      'savings': '₹25 OFF',
      'badge': '11% OFF',
      'rating': '4.5 (8.9k)',
      'barcode': '8901063325036',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/332/5036/front_en.10.400.jpg',
    },
    {
      'name': 'Pintola High Protein Oats',
      'category': 'Breakfast & Sauces',
      'weight': '400g Pack',
      'price': '₹483',
      'mrp': '₹549',
      'savings': '₹66 OFF',
      'badge': '12% OFF',
      'rating': '4.7 (21.5k)',
      'barcode': '8906136651951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.10.400.jpg',
    },
    {
      'name': 'Kellogg\'s Choco Fills Cereal',
      'category': 'Breakfast & Sauces',
      'weight': '250g Pack',
      'price': '₹383',
      'mrp': '₹432',
      'savings': '₹49 OFF',
      'badge': '11% OFF',
      'rating': '4.8 (32.1k)',
      'barcode': '8901058003181',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/149/900/9708/front_en.3.400.jpg',
    },
    {
      'name': 'DoodhShakti Pure Cow Ghee',
      'category': 'Dairy, Bread & Eggs',
      'weight': '500ml Jar',
      'price': '₹866',
      'mrp': '₹992',
      'savings': '₹126 OFF',
      'badge': '13% OFF',
      'rating': '4.6 (11.8k)',
      'barcode': '8901030678912',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/154/200/2090/front_en.5.400.jpg',
    },
    {
      'name': 'Vikram Premium Kadak Dust Tea',
      'category': 'Tea, Coffee & More',
      'weight': '500g Pack',
      'price': '₹113',
      'mrp': '₹121',
      'savings': '₹8 OFF',
      'badge': '6% OFF',
      'rating': '4.4 (5.3k)',
      'barcode': '8901052000155',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/778/0056/front_en.4.400.jpg',
    },
    {
      'name': 'Tata Tea Gold Leaf 500g',
      'category': 'Tea, Coffee & More',
      'weight': '500g Pack',
      'price': '₹315',
      'mrp': '₹350',
      'savings': '₹35 OFF',
      'badge': '10% OFF',
      'rating': '4.7 (19.8k)',
      'barcode': '8901052000155',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
    },
    {
      'name': 'Milk Bikis Crunchy Biscuits',
      'category': 'Packaged Food',
      'weight': '300g Pack',
      'price': '₹79',
      'mrp': '₹89',
      'savings': '₹10 OFF',
      'badge': '11% OFF',
      'rating': '4.7 (45.2k)',
      'barcode': '8901063012011',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2660/front_en.10.400.jpg',
    },
    {
      'name': 'Britannia Good Day Cashew 200g',
      'category': 'Packaged Food',
      'weight': '200g Pack',
      'price': '₹45',
      'mrp': '₹50',
      'savings': '₹5 OFF',
      'badge': '10% OFF',
      'rating': '4.6 (15.3k)',
      'barcode': '8901063012011',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
    {
      'name': 'Amul Taaza Toned Milk 1L',
      'category': 'Dairy, Bread & Eggs',
      'weight': '1L Pack',
      'price': '₹68',
      'mrp': '₹72',
      'savings': '₹4 OFF',
      'badge': '5% OFF',
      'rating': '4.8 (34.1k)',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
    },
    {
      'name': 'Fortune Sunlite Sunflower Oil 1L',
      'category': 'Atta, Rice, Oil & Dals',
      'weight': '1L Pouch',
      'price': '₹142',
      'mrp': '₹165',
      'savings': '₹23 OFF',
      'badge': '14% OFF',
      'rating': '4.6 (12.4k)',
      'barcode': '8906007280015',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
    },
    {
      'name': 'Saffola Active Edible Oil 1L',
      'category': 'Atta, Rice, Oil & Dals',
      'weight': '1L Jar',
      'price': '₹356',
      'mrp': '₹420',
      'savings': '₹64 OFF',
      'badge': '15% OFF',
      'rating': '4.7 (18.2k)',
      'barcode': '8901088034593',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/108/803/4593/front_en.10.400.jpg',
    },
    {
      'name': 'Amul Pasteurised Butter 500g',
      'category': 'Dairy, Bread & Eggs',
      'weight': '500g Pack',
      'price': '₹280',
      'mrp': '₹295',
      'savings': '₹15 OFF',
      'badge': '5% OFF',
      'rating': '4.9 (42.5k)',
      'barcode': '8901030678912',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
    },
    {
      'name': 'Fresh Organic Bananas 1kg',
      'category': 'Fruits & Vegetables',
      'weight': '1kg Pack',
      'price': '₹65',
      'mrp': '₹80',
      'savings': '₹15 OFF',
      'badge': '18% OFF',
      'rating': '4.8 (28.4k)',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
    {
      'name': 'Fresh Farm Eggs (Pack of 6)',
      'category': 'Meat, Fish & Eggs',
      'weight': '6 Eggs Pack',
      'price': '₹55',
      'mrp': '₹62',
      'savings': '₹7 OFF',
      'badge': '11% OFF',
      'rating': '4.7 (12.1k)',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
    },
    {
      'name': 'Catch Turmeric Powder 200g',
      'category': 'Masala & Dry Fruits',
      'weight': '200g Pack',
      'price': '₹62',
      'mrp': '₹70',
      'savings': '₹8 OFF',
      'badge': '11% OFF',
      'rating': '4.6 (8.4k)',
      'barcode': '8906136651951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.10.400.jpg',
    },
    {
      'name': 'Kwality Walls Vanilla Magic 700ml',
      'category': 'Ice Creams & More',
      'weight': '700ml Tub',
      'price': '₹175',
      'mrp': '₹199',
      'savings': '₹24 OFF',
      'badge': '12% OFF',
      'rating': '4.8 (19.2k)',
      'barcode': '8901499009708',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/149/900/9708/front_en.3.400.jpg',
    },
    {
      'name': 'McCain French Fries 400g',
      'category': 'Frozen Food',
      'weight': '400g Pack',
      'price': '₹135',
      'mrp': '₹150',
      'savings': '₹15 OFF',
      'badge': '10% OFF',
      'rating': '4.6 (14.8k)',
      'barcode': '8901542002090',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/154/200/2090/front_en.5.400.jpg',
    },
  ];

  @override
  Widget build(BuildContext context) {
    final catName = widget.categoryName;
    final matchingProducts = (catName == 'All' || catName.isEmpty)
        ? _allProducts
        : _allProducts.where((p) => p['category'] == catName).toList();

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        iconTheme: const IconThemeData(color: Color(0xFF111827)),
        title: Text(
          catName,
          style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 17, color: Color(0xFF111827)),
        ),
        actions: [
          BlocBuilder<CartBloc, CartState>(
            builder: (context, state) {
              int itemCount = 0;
              if (state is CartLoaded) itemCount = state.summary.itemCount;
              if (state is CartCouponError) itemCount = state.summary?.itemCount ?? 0;

              return IconButton(
                onPressed: () => context.push('/cart'),
                icon: Stack(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: const BoxDecoration(
                        color: Color(0xFFF1F5F9),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.shopping_cart_outlined, color: Color(0xFF111827), size: 20),
                    ),
                    if (itemCount > 0)
                      Positioned(
                        right: 0,
                        top: 0,
                        child: Container(
                          padding: const EdgeInsets.all(4),
                          decoration: const BoxDecoration(
                            color: Color(0xFFE02461),
                            shape: BoxShape.circle,
                          ),
                          child: Text(
                            '$itemCount',
                            style: const TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.bold),
                          ),
                        ),
                      ),
                  ],
                ),
              );
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: SafeArea(
        child: Column(
          children: [
            // Category Info Top Header Bar
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: const BoxDecoration(
                color: Colors.white,
                border: Border(bottom: BorderSide(color: Color(0xFFE2E8F0))),
              ),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: const Color(0xFFEFF6FF),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: const Color(0xFFBFDBFE)),
                    ),
                    child: Text(
                      '${matchingProducts.length} Items Available',
                      style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Color(0xFF1E3A8A)),
                    ),
                  ),
                  const Spacer(),
                  const Icon(Icons.tune, size: 16, color: Color(0xFF64748B)),
                  const SizedBox(width: 4),
                  const Text('Filter', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF64748B))),
                ],
              ),
            ),

            // Products Grid View
            Expanded(
              child: matchingProducts.isNotEmpty
                  ? GridView.builder(
                      padding: const EdgeInsets.all(14),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        childAspectRatio: 0.64,
                        crossAxisSpacing: 12,
                        mainAxisSpacing: 12,
                      ),
                      itemCount: matchingProducts.length,
                      itemBuilder: (context, index) {
                        return _buildProductCard(matchingProducts[index]);
                      },
                    )
                  : Center(
                      child: Padding(
                        padding: const EdgeInsets.all(24),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.inventory_2_outlined, size: 64, color: Color(0xFF94A3B8)),
                            const SizedBox(height: 12),
                            Text(
                              'No products found in "$catName"',
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: Color(0xFF1E293B)),
                            ),
                            const SizedBox(height: 6),
                            const Text(
                              'Explore other categories or check back soon!',
                              style: TextStyle(color: Color(0xFF64748B), fontSize: 12),
                            ),
                            const SizedBox(height: 16),
                            ElevatedButton(
                              onPressed: () => context.pop(),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: const Color(0xFF1E3A8A),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                              ),
                              child: const Text('Back to Categories', style: TextStyle(color: Colors.white)),
                            ),
                          ],
                        ),
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildProductCard(Map<String, dynamic> item) {
    final name = item['name'] as String;
    final weight = item['weight'] as String;
    final price = item['price'] as String;
    final mrp = item['mrp'] as String? ?? '';
    final badge = item['badge'] as String? ?? '';
    final rating = item['rating'] as String? ?? '4.7';
    final imageUrl = item['image_url'] as String;
    final barcode = item['barcode'] as String? ?? '8904335601951';

    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE5E7EB)),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.03), blurRadius: 8, offset: const Offset(0, 2)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Badge
          if (badge.isNotEmpty)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: const Color(0xFF2563EB),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                badge,
                style: const TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.w900),
              ),
            ),
          const SizedBox(height: 6),

          // Image
          Expanded(
            child: Center(
              child: Image.network(
                imageUrl,
                fit: BoxFit.contain,
                errorBuilder: (_, __, ___) => const Icon(Icons.inventory_2_outlined, size: 40, color: Colors.grey),
              ),
            ),
          ),
          const SizedBox(height: 8),

          // Rating
          Row(
            children: [
              const Icon(Icons.star_rounded, color: Color(0xFF16A34A), size: 14),
              const SizedBox(width: 2),
              Text(rating, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Color(0xFF4B5563))),
            ],
          ),
          const SizedBox(height: 4),

          // Title
          Text(
            name,
            style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF1F2937), height: 1.2),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 2),
          Text(weight, style: const TextStyle(fontSize: 10, color: Color(0xFF6B7280))),
          const SizedBox(height: 8),

          // Price & ADD button
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(price, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w900, color: Color(0xFF111827))),
                  if (mrp.isNotEmpty)
                    Text(
                      mrp,
                      style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF), decoration: TextDecoration.lineThrough),
                    ),
                ],
              ),
              GestureDetector(
                onTap: () {
                  context.push('/scan', extra: {
                    'name': name,
                    'barcode': barcode,
                    'price': price,
                    'image_url': imageUrl,
                  });
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    border: Border.all(color: const Color(0xFFE02461), width: 1.5),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Text(
                    'ADD',
                    style: TextStyle(color: Color(0xFFE02461), fontSize: 11, fontWeight: FontWeight.w900),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

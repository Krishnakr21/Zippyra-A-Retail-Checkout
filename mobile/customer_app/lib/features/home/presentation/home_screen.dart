import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import '../../../injection_container.dart';
import '../../cart/presentation/bloc/cart_bloc.dart';
import 'bloc/home_bloc.dart';
import 'bloc/home_event.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<HomeBloc>()..add(const LoadHomeDataEvent()),
      child: const _HomeScreenContent(),
    );
  }
}

class _HomeScreenContent extends StatefulWidget {
  const _HomeScreenContent();

  @override
  State<_HomeScreenContent> createState() => _HomeScreenContentState();
}

class _HomeScreenContentState extends State<_HomeScreenContent> {
  int _selectedCategoryIndex = 0;

  final List<Map<String, dynamic>> _categoriesList = const [
    {
      'name': 'All',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
    },
    {
      'name': 'Fruits & Vegetables',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
    {
      'name': 'Dairy, Bread & Eggs',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
    },
    {
      'name': 'Atta, Rice, Oil & Dals',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
    },
    {
      'name': 'Meat, Fish & Eggs',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
    },
    {
      'name': 'Masala & Dry Fruits',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.14.400.jpg',
    },
    {
      'name': 'Breakfast & Sauces',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
    },
    {
      'name': 'Packaged Food',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/332/5036/front_en.10.400.jpg',
    },
    {
      'name': 'Tea, Coffee & More',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
    },
    {
      'name': 'Ice Creams & More',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/149/900/9708/front_en.3.400.jpg',
    },
    {
      'name': 'Frozen Food',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/154/200/2090/front_en.5.400.jpg',
    },
  ];

  List<Map<String, dynamic>> get _filteredProducts {
    if (_selectedCategoryIndex == 0) return _products;
    final selectedCatName = _categoriesList[_selectedCategoryIndex]['name'] as String;
    return _products.where((p) => p['category'] == selectedCatName).toList();
  }

  // Dedicated previous order items for Buy Again section
  final List<Map<String, dynamic>> _previousOrderItems = [
    {
      'name': 'Amul Taaza Toned Milk 1L',
      'barcode': '8901262010054',
      'price': '₹68',
      'mrp': '₹72',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
      'subtitle': 'Ordered 2 days ago',
    },
    {
      'name': 'Tata Tea Gold Leaf 500g',
      'barcode': '8901052000155',
      'price': '₹315',
      'mrp': '₹350',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
      'subtitle': 'Ordered 4 days ago',
    },
    {
      'name': 'Britannia Good Day Cashew 200g',
      'barcode': '8901063012011',
      'price': '₹45',
      'mrp': '₹50',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
      'subtitle': 'Ordered 1 week ago',
    },
    {
      'name': 'Fortune Sunlite Sunflower Oil 1L',
      'barcode': '8906007280015',
      'price': '₹142',
      'mrp': '₹165',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
      'subtitle': 'Ordered 2 weeks ago',
    },
  ];

  // Preseeded Zepto-style top products categorized for instant filtering
  List<Map<String, dynamic>> _products = [
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
  ];

  @override
  void initState() {
    super.initState();
    _fetchCatalogBackendData();
  }

  Future<void> _fetchCatalogBackendData() async {
    try {
      final dio = Dio();
      final response = await dio.get('http://localhost:8083/v1/catalog/search?store_id=store-1');
      if (response.statusCode == 200 && response.data != null) {
        final productsList = response.data['products'] as List?;
        if (productsList != null && productsList.isNotEmpty) {
          setState(() {
            _products = productsList.map((p) {
              final pricePaise = (p['price_paise'] ?? 0) as int;
              final mrpPaise = (p['mrp_paise'] ?? 0) as int;
              final priceRs = (pricePaise / 100.0).round();
              final mrpRs = (mrpPaise / 100.0).round();
              final isDiscounted = mrpPaise > pricePaise;
              final discountPercent = isDiscounted
                  ? (((mrpPaise - pricePaise) / mrpPaise) * 100).round()
                  : 0;
              final savingsRs = isDiscounted ? ((mrpPaise - pricePaise) / 100.0).round() : 0;
              final barcode = p['barcode'] as String? ?? '0';

              final ratingVal = '4.${5 + (barcode.hashCode % 4)} (${1 + (barcode.hashCode % 15)}.${(barcode.hashCode % 9)}k)';

              return {
                'name': p['name'] ?? 'Product',
                'weight': 'Fresh Stock',
                'price': '₹$priceRs',
                'mrp': isDiscounted ? '₹$mrpRs' : '',
                'savings': savingsRs > 0 ? '₹$savingsRs OFF' : '',
                'badge': isDiscounted ? '$discountPercent% OFF' : 'Bestseller',
                'rating': ratingVal,
                'barcode': barcode,
                'image_url': p['image_url'] ?? '',
              };
            }).toList();
          });
        }
      }
    } catch (_) {
      // Retain formatted preseeded list if offline
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF3F4F6),
      body: SafeArea(
        child: Column(
          children: [
            // Top Location & Cart Header
            Container(
              color: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(6),
                            decoration: BoxDecoration(
                              color: const Color(0xFFEFF6FF),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(Icons.location_on_rounded, color: Color(0xFF2563EB), size: 18),
                          ),
                          const SizedBox(width: 8),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: const [
                              Row(
                                children: [
                                  Text(
                                    'Smart Bazaar',
                                    style: TextStyle(fontWeight: FontWeight.w900, fontSize: 14, color: Color(0xFF111827)),
                                  ),
                                  Icon(Icons.keyboard_arrow_down_rounded, color: Color(0xFF111827), size: 18),
                                ],
                              ),
                              Text(
                                'Koramangala, Bengaluru',
                                style: TextStyle(fontSize: 10, color: Color(0xFF6B7280), fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                        ],
                      ),
                      Row(
                        children: [
                          const SizedBox(width: 8),

                          // Cart Icon with Dynamic Counter
                          BlocBuilder<CartBloc, CartState>(
                            builder: (context, state) {
                              int itemCount = 0;
                              if (state is CartLoaded) {
                                itemCount = state.summary.itemCount;
                              } else if (state is CartCouponError) {
                                itemCount = state.summary?.itemCount ?? 0;
                              }

                              return GestureDetector(
                                onTap: () => context.push('/cart'),
                                child: Stack(
                                  clipBehavior: Clip.none,
                                  children: [
                                    Container(
                                      width: 36,
                                      height: 36,
                                      decoration: BoxDecoration(
                                        color: const Color(0xFFF3F4F6),
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      child: const Icon(Icons.shopping_cart_outlined, size: 20, color: Color(0xFF111827)),
                                    ),
                                    if (itemCount > 0)
                                      Positioned(
                                        top: -2,
                                        right: -2,
                                        child: Container(
                                          width: 16,
                                          height: 16,
                                          decoration: const BoxDecoration(
                                            color: Color(0xFFE02461),
                                            shape: BoxShape.circle,
                                          ),
                                          child: Center(
                                            child: Text(
                                              '$itemCount',
                                              style: const TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.w900),
                                            ),
                                          ),
                                        ),
                                      ),
                                  ],
                                ),
                              );
                            },
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),

                  // Search Bar Trigger
                  GestureDetector(
                    onTap: () => context.push('/search'),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF3F4F6),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFE5E7EB)),
                      ),
                      child: Row(
                        children: const [
                          Icon(Icons.search_rounded, color: Color(0xFF9CA3AF), size: 20),
                          SizedBox(width: 10),
                          Text(
                            'Search "milk", "oats", "tea", "biscuits"…',
                            style: TextStyle(color: Color(0xFF6B7280), fontSize: 13, fontWeight: FontWeight.w500),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Zepto Category Cards Bar (Matching Screenshot 2)
            Container(
              color: Colors.white,
              height: 106,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                itemCount: _categoriesList.length,
                itemBuilder: (context, index) {
                  final item = _categoriesList[index];
                  final isSelected = index == _selectedCategoryIndex;

                  return GestureDetector(
                    onTap: () {
                      if (index == 0) {
                        setState(() => _selectedCategoryIndex = 0);
                      } else {
                        context.push('/category/products?name=${Uri.encodeComponent(item['name'] as String)}');
                      }
                    },
                    child: Container(
                      width: 82,
                      margin: const EdgeInsets.only(right: 8),
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: isSelected ? const Color(0xFFEFF6FF) : const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(
                          color: isSelected ? const Color(0xFF2563EB) : const Color(0xFFE2E8F0),
                          width: isSelected ? 2 : 1,
                        ),
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Container(
                            width: 52,
                            height: 52,
                            padding: const EdgeInsets.all(4),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: Image.network(
                              item['image_url'] as String,
                              fit: BoxFit.contain,
                              errorBuilder: (_, __, ___) => const Icon(Icons.category_outlined, size: 24, color: Colors.grey),
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            item['name'] as String,
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 9,
                              fontWeight: isSelected ? FontWeight.w900 : FontWeight.w700,
                              color: isSelected ? const Color(0xFF1E3A8A) : const Color(0xFF334155),
                              height: 1.1,
                            ),
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
            const Divider(height: 1, color: Color(0xFFE5E7EB)),

            // Main Scrollable Body
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(14),
                physics: const BouncingScrollPhysics(),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Zepto Scan & Go Banner
                    GestureDetector(
                      onTap: () => context.push('/scan'),
                      child: Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(18),
                          gradient: const LinearGradient(
                            colors: [Color(0xFF1E3A8A), Color(0xFF0F172A)],
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                          ),
                          boxShadow: [
                            BoxShadow(
                              color: const Color(0xFF1E3A8A).withOpacity(0.3),
                              blurRadius: 12,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Row(
                                children: const [
                                  Icon(Icons.bolt_rounded, color: Color(0xFFFBBF24), size: 20),
                                  SizedBox(width: 4),
                                  Text(
                                    'SCAN & GO CHECKOUT',
                                    style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w900, letterSpacing: 0.5),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(
                                'Scan item barcodes & skip long store queues',
                                style: TextStyle(color: Colors.white.withOpacity(0.8), fontSize: 11, fontWeight: FontWeight.w500),
                              ),
                              const SizedBox(height: 10),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
                                decoration: BoxDecoration(
                                  color: Colors.white.withOpacity(0.2),
                                  borderRadius: BorderRadius.circular(8),
                                  border: Border.all(color: Colors.white.withOpacity(0.3)),
                                ),
                                child: const Text(
                                  'Launch Scanner →',
                                  style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                                ),
                              ),
                            ],
                          ),
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.15),
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(Icons.qr_code_scanner_rounded, color: Colors.white, size: 36),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 14),

                    // In-Store Exit Gate Ready Badge
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(color: const Color(0xFFE5E7EB)),
                      ),
                      child: Row(
                        children: [
                          Container(
                            width: 36,
                            height: 36,
                            decoration: BoxDecoration(
                              color: const Color(0xFFDCFCE7),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: const Icon(Icons.sensor_door_rounded, color: Color(0xFF16A34A), size: 20),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: const [
                                Text(
                                  'Exit Gate Verification Active',
                                  style: TextStyle(fontSize: 12, fontWeight: FontWeight.w800, color: Color(0xFF111827)),
                                ),
                                Text(
                                  'CV cameras & RFID automated validation',
                                  style: TextStyle(fontSize: 10, color: Color(0xFF6B7280), fontWeight: FontWeight.w500),
                                ),
                              ],
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                            decoration: BoxDecoration(
                              color: const Color(0xFF16A34A),
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: const Text(
                              'READY',
                              style: TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.w900),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 18),

                    // Buy Again Section
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: const [
                        Text(
                          'Buy Again',
                          style: TextStyle(fontSize: 15, fontWeight: FontWeight.w900, color: Color(0xFF111827)),
                        ),
                        Text(
                          'See All',
                          style: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Color(0xFF2563EB)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 10),

                    SizedBox(
                      height: 140,
                      child: ListView.builder(
                        scrollDirection: Axis.horizontal,
                        physics: const BouncingScrollPhysics(),
                        itemCount: _previousOrderItems.length,
                        itemBuilder: (context, index) {
                          final item = _previousOrderItems[index];
                          return Container(
                            width: 110,
                            margin: const EdgeInsets.only(right: 10),
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: const Color(0xFFE5E7EB)),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Expanded(
                                  child: Center(
                                    child: Image.network(
                                      item['image_url'] as String,
                                      fit: BoxFit.contain,
                                      errorBuilder: (_, __, ___) => const Icon(Icons.inventory_2_outlined, size: 30, color: Colors.grey),
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  item['name'] as String,
                                  style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: Color(0xFF1F2937)),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    Text(
                                      item['price'] as String,
                                      style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w900, color: Color(0xFF111827)),
                                    ),
                                    GestureDetector(
                                      onTap: () {
                                        context.push('/scan', extra: {
                                          'name': item['name'] as String,
                                          'barcode': item['barcode'] as String? ?? '8904335601951',
                                          'price': item['price'] as String,
                                          'image_url': item['image_url'] as String,
                                        });
                                      },
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                                        decoration: BoxDecoration(
                                          color: Colors.white,
                                          border: Border.all(color: const Color(0xFFE02461), width: 1.2),
                                          borderRadius: BorderRadius.circular(6),
                                        ),
                                        child: const Text(
                                          'ADD',
                                          style: TextStyle(color: Color(0xFFE02461), fontSize: 9, fontWeight: FontWeight.w900),
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Top Picks Product Grid (Zepto Style)
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: const [
                        Text(
                          'Recommended Products',
                          style: TextStyle(fontSize: 15, fontWeight: FontWeight.w900, color: Color(0xFF111827)),
                        ),
                        Text(
                          'View All',
                          style: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Color(0xFF2563EB)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),

                    _filteredProducts.isNotEmpty
                        ? GridView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 2,
                              childAspectRatio: 0.62,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                            ),
                            itemCount: _filteredProducts.length,
                            itemBuilder: (context, index) {
                              return _buildZeptoProductCard(_filteredProducts[index]);
                            },
                          )
                        : Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(24),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(color: const Color(0xFFE2E8F0)),
                            ),
                            child: Column(
                              children: [
                                const Icon(Icons.shopping_bag_outlined, color: Color(0xFF94A3B8), size: 48),
                                const SizedBox(height: 8),
                                const Text(
                                  'No items in this category',
                                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E293B)),
                                ),
                                const SizedBox(height: 4),
                                const Text(
                                  'Select "All" above to explore all store items',
                                  style: TextStyle(fontSize: 11, color: Color(0xFF64748B)),
                                ),
                                const SizedBox(height: 12),
                                ElevatedButton(
                                  onPressed: () => setState(() => _selectedCategoryIndex = 0),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFF1E3A8A),
                                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                                  ),
                                  child: const Text('Show All Products', style: TextStyle(color: Colors.white, fontSize: 12)),
                                ),
                              ],
                            ),
                          ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: Container(
        height: 64,
        decoration: BoxDecoration(
          color: Colors.white,
          border: const Border(top: BorderSide(color: Color(0xFFE5E7EB))),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 10,
              offset: const Offset(0, -4),
            ),
          ],
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildNavItem(icon: Icons.home_rounded, label: 'Home', isActive: true, onTap: () {}),
            _buildNavItem(icon: Icons.grid_view_rounded, label: 'Categ.', isActive: false, onTap: () => context.push('/categories')),
            GestureDetector(
              onTap: () => context.push('/scan'),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Transform.translate(
                    offset: const Offset(0, -6),
                    child: Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(16),
                        gradient: const LinearGradient(
                          colors: [Color(0xFF1E3A8A), Color(0xFF0F172A)],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: const Color(0xFF1E3A8A).withOpacity(0.3),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: const Center(
                        child: Icon(Icons.qr_code_scanner_rounded, color: Colors.white, size: 24),
                      ),
                    ),
                  ),
                  const Text(
                    'Scan',
                    style: TextStyle(
                      fontSize: 9,
                      fontWeight: FontWeight.w800,
                      color: Color(0xFF1E3A8A),
                    ),
                  ),
                ],
              ),
            ),
            _buildNavItem(icon: Icons.inventory_2_outlined, label: 'Orders', isActive: false, onTap: () => context.push('/orders')),
            _buildNavItem(icon: Icons.person_outline_rounded, label: 'Profile', isActive: false, onTap: () => context.push('/profile')),
          ],
        ),
      ),
    );
  }

  Widget _buildZeptoProductCard(Map<String, dynamic> product) {
    final name = product['name'] as String? ?? 'Product';
    final price = product['price'] as String? ?? '₹0';
    final mrp = product['mrp'] as String? ?? '';
    final badge = product['badge'] as String? ?? '';
    final weight = product['weight'] as String? ?? '1 pack';
    final rating = product['rating'] as String? ?? '4.5 (2.1k)';
    final imageUrl = product['image_url'] as String? ?? '';
    final savings = product['savings'] as String? ?? '';

    final bool hasBadge = badge.isNotEmpty;

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
      child: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(10.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Image Frame
                Expanded(
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF9FAFB),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Center(
                      child: imageUrl.isNotEmpty
                          ? Image.network(
                              imageUrl,
                              fit: BoxFit.contain,
                              errorBuilder: (_, __, ___) => const Icon(
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

                // Star Rating Tag
                Row(
                  children: [
                    const Icon(Icons.star_rounded, color: Color(0xFF16A34A), size: 14),
                    const SizedBox(width: 2),
                    Text(
                      rating,
                      style: const TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF4B5563),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),

                // Product Title (Max 2 lines)
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1F2937),
                    height: 1.2,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),

                // Pack Size / Weight
                Text(
                  weight,
                  style: const TextStyle(
                    fontSize: 10,
                    color: Color(0xFF6B7280),
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),

                // Price & Zepto ADD Button Row
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
                              price,
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w900,
                                color: Color(0xFF111827),
                              ),
                            ),
                            if (mrp.isNotEmpty) ...[
                              const SizedBox(width: 4),
                              Text(
                                mrp,
                                style: const TextStyle(
                                  fontSize: 10,
                                  color: Color(0xFF9CA3AF),
                                  decoration: TextDecoration.lineThrough,
                                ),
                              ),
                            ],
                          ],
                        ),
                        if (savings.isNotEmpty) ...[
                          const SizedBox(height: 1),
                          Text(
                            savings,
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
                      onTap: () {
                        context.push('/scan', extra: {
                          'name': name,
                          'barcode': product['barcode'] as String? ?? '8904335601951',
                          'price': price,
                          'image_url': imageUrl,
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
          if (hasBadge)
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
                  badge,
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
    );
  }

  Widget _buildNavItem({
    required IconData icon,
    required String label,
    required bool isActive,
    required VoidCallback onTap,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            icon,
            size: 20,
            color: isActive ? const Color(0xFF1E3A8A) : const Color(0xFF6B7280),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              fontSize: 9,
              fontWeight: FontWeight.bold,
              color: isActive ? const Color(0xFF1E3A8A) : const Color(0xFF6B7280),
            ),
          ),
        ],
      ),
    );
  }
}

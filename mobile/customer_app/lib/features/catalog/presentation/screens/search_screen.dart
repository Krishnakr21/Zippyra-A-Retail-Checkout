import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import '../../../cart/presentation/bloc/cart_bloc.dart';

class SearchScreen extends StatefulWidget {
  final String storeId;

  const SearchScreen({super.key, this.storeId = 'store-1'});

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final TextEditingController _searchController = TextEditingController();
  String _searchQuery = '';
  String _activeFilter = 'All';

  final List<String> _recentSearches = ['Milk', 'Amul Butter', 'Bread', 'Eggs', 'Tea'];

  final List<Map<String, String>> _trendingItems = [
    {
      'title': 'Real Juice Mango',
      'tag': 'Trending',
      'tagColor': '0xFFFFEDD5',
      'textColor': '0xFFC2410C',
      'icon': '🧃'
    },
    {
      'title': 'Cadbury Dairy Milk',
      'tag': 'Hot',
      'tagColor': '0xFFFEE2E2',
      'textColor': '0xFFB91C1C',
      'icon': '🍫'
    },
    {
      'title': 'Red Bull Energy Drink',
      'tag': '+40% Today',
      'tagColor': '0xFFDCFCE7',
      'textColor': '0xFF15803D',
      'icon': '🥤'
    },
    {
      'title': 'Amul Taaza Toned Milk',
      'tag': 'Popular',
      'tagColor': '0xFFDBEAFE',
      'textColor': '0xFF1D4ED8',
      'icon': '🥛'
    },
  ];

  final List<Map<String, dynamic>> _catalogDatabase = const [
    {
      'name': 'Amul Pasteurised Butter 500g',
      'aisle': 'Aisle B3',
      'category': 'Dairy, Bread & Eggs',
      'weight': '500g Pack',
      'price': '₹280',
      'mrp': '₹295',
      'badge': '10% OFF',
      'rating': '4.9 (42.5k)',
      'barcode': '8901030678912',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
    },
    {
      'name': 'Amul Taaza Toned Milk 1L',
      'aisle': 'Aisle D1',
      'category': 'Dairy, Bread & Eggs',
      'weight': '1L Pack',
      'price': '₹68',
      'mrp': '₹72',
      'badge': '5% OFF',
      'rating': '4.8 (34.1k)',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
    },
    {
      'name': '100% Rolled Oats (Yoga Bar)',
      'aisle': 'Aisle C2',
      'category': 'Breakfast & Sauces',
      'weight': '500g Pack',
      'price': '₹493',
      'mrp': '₹530',
      'badge': '7% OFF',
      'rating': '4.6 (14.2k)',
      'barcode': '8904335601951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
    },
    {
      'name': 'Britannia Toastea Bake Rusk',
      'aisle': 'Aisle A4',
      'category': 'Packaged Food',
      'weight': '250g Pack',
      'price': '₹212',
      'mrp': '₹237',
      'badge': '11% OFF',
      'rating': '4.5 (8.9k)',
      'barcode': '8901063325036',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/332/5036/front_en.10.400.jpg',
    },
    {
      'name': 'Pintola High Protein Oats',
      'aisle': 'Aisle C2',
      'category': 'Breakfast & Sauces',
      'weight': '400g Pack',
      'price': '₹483',
      'mrp': '₹549',
      'badge': '12% OFF',
      'rating': '4.7 (21.5k)',
      'barcode': '8906136651951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.10.400.jpg',
    },
    {
      'name': 'Kellogg\'s Choco Fills Cereal',
      'aisle': 'Aisle C1',
      'category': 'Breakfast & Sauces',
      'weight': '250g Pack',
      'price': '₹383',
      'mrp': '₹432',
      'badge': '11% OFF',
      'rating': '4.8 (32.1k)',
      'barcode': '8901058003181',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/149/900/9708/front_en.3.400.jpg',
    },
    {
      'name': 'DoodhShakti Pure Cow Ghee',
      'aisle': 'Aisle B1',
      'category': 'Dairy, Bread & Eggs',
      'weight': '500ml Jar',
      'price': '₹866',
      'mrp': '₹992',
      'badge': '13% OFF',
      'rating': '4.6 (11.8k)',
      'barcode': '8901030678912',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/154/200/2090/front_en.5.400.jpg',
    },
    {
      'name': 'Vikram Premium Kadak Dust Tea',
      'aisle': 'Aisle E3',
      'category': 'Tea, Coffee & More',
      'weight': '500g Pack',
      'price': '₹113',
      'mrp': '₹121',
      'badge': '6% OFF',
      'rating': '4.4 (5.3k)',
      'barcode': '8901052000155',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/778/0056/front_en.4.400.jpg',
    },
    {
      'name': 'Tata Tea Gold Leaf 500g',
      'aisle': 'Aisle E3',
      'category': 'Tea, Coffee & More',
      'weight': '500g Pack',
      'price': '₹315',
      'mrp': '₹350',
      'badge': '10% OFF',
      'rating': '4.7 (19.8k)',
      'barcode': '8901052000155',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
    },
    {
      'name': 'Milk Bikis Crunchy Biscuits',
      'aisle': 'Aisle A2',
      'category': 'Packaged Food',
      'weight': '300g Pack',
      'price': '₹79',
      'mrp': '₹89',
      'badge': '11% OFF',
      'rating': '4.7 (45.2k)',
      'barcode': '8901063012011',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2660/front_en.10.400.jpg',
    },
    {
      'name': 'Britannia Good Day Cashew 200g',
      'aisle': 'Aisle A2',
      'category': 'Packaged Food',
      'weight': '200g Pack',
      'price': '₹45',
      'mrp': '₹50',
      'badge': '10% OFF',
      'rating': '4.6 (15.3k)',
      'barcode': '8901063012011',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
    {
      'name': 'Fortune Sunlite Sunflower Oil 1L',
      'aisle': 'Aisle B2',
      'category': 'Atta, Rice, Oil & Dals',
      'weight': '1L Pouch',
      'price': '₹142',
      'mrp': '₹165',
      'badge': '14% OFF',
      'rating': '4.6 (12.4k)',
      'barcode': '8906007280015',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
    },
    {
      'name': 'Saffola Active Edible Oil 1L',
      'aisle': 'Aisle B2',
      'category': 'Atta, Rice, Oil & Dals',
      'weight': '1L Jar',
      'price': '₹356',
      'mrp': '₹420',
      'badge': '15% OFF',
      'rating': '4.7 (18.2k)',
      'barcode': '8901088034593',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/108/803/4593/front_en.10.400.jpg',
    },
    {
      'name': 'Fresh Organic Bananas 1kg',
      'aisle': 'Aisle F1',
      'category': 'Fruits & Vegetables',
      'weight': '1kg Pack',
      'price': '₹65',
      'mrp': '₹80',
      'badge': '18% OFF',
      'rating': '4.8 (28.4k)',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
  ];

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<Map<String, dynamic>> get _filteredResults {
    final query = _searchQuery.trim().toLowerCase();
    if (query.isEmpty) return [];

    return _catalogDatabase.where((item) {
      final name = (item['name'] as String).toLowerCase();
      final category = (item['category'] as String).toLowerCase();
      final aisle = (item['aisle'] as String).toLowerCase();
      final barcode = (item['barcode'] as String).toLowerCase();

      final matchesQuery = name.contains(query) || category.contains(query) || aisle.contains(query) || barcode.contains(query);

      if (!matchesQuery) return false;

      if (_activeFilter == 'Offers') {
        return (item['badge'] as String? ?? '').isNotEmpty;
      } else if (_activeFilter == 'Under ₹200') {
        final priceNum = int.tryParse((item['price'] as String).replaceAll('₹', '')) ?? 0;
        return priceNum < 200;
      }

      return true;
    }).toList();
  }

  void _executeSearch(String text) {
    setState(() {
      _searchQuery = text;
      _searchController.text = text;
      _searchController.selection = TextSelection.fromPosition(TextPosition(offset: text.length));
    });
  }

  @override
  Widget build(BuildContext context) {
    final results = _filteredResults;
    final isQueryActive = _searchQuery.trim().isNotEmpty;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Color(0xFF1E293B)),
          onPressed: () => context.pop(),
        ),
        titleSpacing: 0,
        title: Container(
          height: 44,
          margin: const EdgeInsets.only(right: 12),
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFE2E8F0)),
          ),
          child: Row(
            children: [
              const Icon(Icons.search, color: Color(0xFF64748B), size: 20),
              const SizedBox(width: 8),
              Expanded(
                child: TextField(
                  controller: _searchController,
                  autofocus: true,
                  style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF0F172A)),
                  decoration: const InputDecoration(
                    hintText: 'Search "milk", "oats", "tea", "butter"...',
                    hintStyle: TextStyle(color: Color(0xFF94A3B8), fontSize: 13),
                    border: InputBorder.none,
                    isDense: true,
                  ),
                  onChanged: (val) => setState(() => _searchQuery = val),
                ),
              ),
              if (_searchQuery.isNotEmpty)
                GestureDetector(
                  onTap: () {
                    _searchController.clear();
                    setState(() => _searchQuery = '');
                  },
                  child: const Icon(Icons.close_rounded, color: Color(0xFF64748B), size: 18),
                ),
            ],
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.qr_code_scanner, color: Color(0xFF1E3A8A)),
            tooltip: 'Scan Barcode',
            onPressed: () => context.push('/scan'),
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: SafeArea(
        child: !isQueryActive
            ? _buildIdleState()
            : Column(
                children: [
                  // Results Summary & Sub-filter chips (Figma Screen 7.4)
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: const BoxDecoration(
                      color: Colors.white,
                      border: Border(bottom: BorderSide(color: Color(0xFFE2E8F0))),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${results.length} results for "$_searchQuery"',
                          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF64748B)),
                        ),
                        const SizedBox(height: 8),
                        SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            children: ['All', 'Offers', 'Under ₹200'].map((filter) {
                              final isSelected = filter == _activeFilter;
                              return GestureDetector(
                                onTap: () => setState(() => _activeFilter = filter),
                                child: Container(
                                  margin: const EdgeInsets.only(right: 8),
                                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
                                  decoration: BoxDecoration(
                                    color: isSelected ? const Color(0xFF1E3A8A) : const Color(0xFFF1F5F9),
                                    borderRadius: BorderRadius.circular(20),
                                    border: Border.all(color: isSelected ? const Color(0xFF1E3A8A) : const Color(0xFFCBD5E1)),
                                  ),
                                  child: Text(
                                    filter,
                                    style: TextStyle(
                                      color: isSelected ? Colors.white : const Color(0xFF334155),
                                      fontWeight: FontWeight.bold,
                                      fontSize: 11,
                                    ),
                                  ),
                                ),
                              );
                            }).toList(),
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Search Results Product Grid (Figma Screen 7.4)
                  Expanded(
                    child: results.isNotEmpty
                        ? GridView.builder(
                            padding: const EdgeInsets.all(14),
                            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 2,
                              childAspectRatio: 0.64,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                            ),
                            itemCount: results.length,
                            itemBuilder: (context, index) {
                              return _buildProductCard(results[index]);
                            },
                          )
                        : Center(
                            child: Padding(
                              padding: const EdgeInsets.all(24),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  const Icon(Icons.search_off_rounded, size: 64, color: Color(0xFF94A3B8)),
                                  const SizedBox(height: 12),
                                  Text(
                                    'No matching products found for "$_searchQuery"',
                                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF1E293B)),
                                  ),
                                  const SizedBox(height: 6),
                                  const Text(
                                    'Try searching "milk", "butter", "oats", or scan product barcode',
                                    style: TextStyle(fontSize: 11, color: Color(0xFF64748B)),
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

  Widget _buildIdleState() {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // RECENT SEARCHES (Figma Screen 7.2)
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: const [
            Text('RECENT SEARCHES', style: TextStyle(color: Color(0xFF64748B), fontSize: 11, fontWeight: FontWeight.w800, letterSpacing: 0.5)),
            Text('Clear All', style: TextStyle(color: Color(0xFF2563EB), fontSize: 11, fontWeight: FontWeight.bold)),
          ],
        ),
        const SizedBox(height: 10),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: _recentSearches.map((term) {
            return GestureDetector(
              onTap: () => _executeSearch(term),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                  boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 2)],
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.history_rounded, size: 14, color: Color(0xFF64748B)),
                    const SizedBox(width: 6),
                    Text(term, style: const TextStyle(color: Color(0xFF1E293B), fontSize: 12, fontWeight: FontWeight.w600)),
                  ],
                ),
              ),
            );
          }).toList(),
        ),

        const SizedBox(height: 28),

        // TRENDING NOW 🔥 (Figma Screen 7.2)
        Row(
          children: const [
            Text('TRENDING NOW', style: TextStyle(color: Color(0xFF64748B), fontSize: 11, fontWeight: FontWeight.w800, letterSpacing: 0.5)),
            SizedBox(width: 4),
            Text('🔥', style: TextStyle(fontSize: 12)),
          ],
        ),
        const SizedBox(height: 12),
        Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFE2E8F0)),
          ),
          child: Column(
            children: _trendingItems.asMap().entries.map((entry) {
              final idx = entry.key;
              final item = entry.value;
              final tagBg = Color(int.parse(item['tagColor']!));
              final tagText = Color(int.parse(item['textColor']!));

              return Column(
                children: [
                  ListTile(
                    leading: Text(item['icon']!, style: const TextStyle(fontSize: 22)),
                    title: Text(
                      item['title']!,
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF1E293B)),
                    ),
                    trailing: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: tagBg,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        item['tag']!,
                        style: TextStyle(color: tagText, fontSize: 10, fontWeight: FontWeight.w800),
                      ),
                    ),
                    onTap: () => _executeSearch(item['title']!),
                  ),
                  if (idx < _trendingItems.length - 1)
                    const Divider(height: 1, indent: 16, endIndent: 16, color: Color(0xFFF1F5F9)),
                ],
              );
            }).toList(),
          ),
        ),
      ],
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
          Row(
            children: [
              const Icon(Icons.star_rounded, color: Color(0xFF16A34A), size: 14),
              const SizedBox(width: 2),
              Text(rating, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Color(0xFF4B5563))),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            name,
            style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF1F2937), height: 1.2),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 2),
          Text(weight, style: const TextStyle(fontSize: 10, color: Color(0xFF6B7280))),
          const SizedBox(height: 8),
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

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class CategoryBrowseScreen extends StatelessWidget {
  final String chainId;

  const CategoryBrowseScreen({super.key, this.chainId = 'chain-hq-001'});

  static const List<Map<String, dynamic>> categoriesList = [
    {
      'name': 'Fruits &\nVegetables',
      'rawName': 'Fruits & Vegetables',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
      'bgColor': Color(0xFFF0FDF4),
      'borderColor': Color(0xFFBBF7D0),
    },
    {
      'name': 'Dairy, Bread\n& Eggs',
      'rawName': 'Dairy, Bread & Eggs',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
      'bgColor': Color(0xFFEFF6FF),
      'borderColor': Color(0xFFBFDBFE),
    },
    {
      'name': 'Atta, Rice,\nOil & Dals',
      'rawName': 'Atta, Rice, Oil & Dals',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
      'bgColor': Color(0xFFFFFBEB),
      'borderColor': Color(0xFFFDE68A),
    },
    {
      'name': 'Meat, Fish\n& Eggs',
      'rawName': 'Meat, Fish & Eggs',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
      'bgColor': Color(0xFFFEF2F2),
      'borderColor': Color(0xFFFECACA),
    },
    {
      'name': 'Masala &\nDry Fruits',
      'rawName': 'Masala & Dry Fruits',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.10.400.jpg',
      'bgColor': Color(0xFFFAF5FF),
      'borderColor': Color(0xFFE9D5FF),
    },
    {
      'name': 'Breakfast &\nSauces',
      'rawName': 'Breakfast & Sauces',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
      'bgColor': Color(0xFFFFF1F2),
      'borderColor': Color(0xFFFECDD3),
    },
    {
      'name': 'Packaged\nFood',
      'rawName': 'Packaged Food',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/332/5036/front_en.10.400.jpg',
      'bgColor': Color(0xFFF0FDFA),
      'borderColor': Color(0xFF99F6E4),
    },
    {
      'name': 'Tea, Coffee\n& More',
      'rawName': 'Tea, Coffee & More',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
      'bgColor': Color(0xFFF5F3FF),
      'borderColor': Color(0xFFDDD6FE),
    },
    {
      'name': 'Ice Creams\n& More',
      'rawName': 'Ice Creams & More',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/149/900/9708/front_en.3.400.jpg',
      'bgColor': Color(0xFFFDF2F8),
      'borderColor': Color(0xFFFBCFE8),
    },
    {
      'name': 'Frozen\nFood',
      'rawName': 'Frozen Food',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/154/200/2090/front_en.5.400.jpg',
      'bgColor': Color(0xFFECFEFF),
      'borderColor': Color(0xFFA5F3FC),
    },
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        iconTheme: const IconThemeData(color: Color(0xFF111827)),
        title: const Text(
          'Explore Categories',
          style: TextStyle(fontWeight: FontWeight.w900, fontSize: 18, color: Color(0xFF111827)),
        ),
      ),
      body: SafeArea(
        child: GridView.builder(
          padding: const EdgeInsets.all(14),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 3,
            childAspectRatio: 0.88,
            crossAxisSpacing: 10,
            mainAxisSpacing: 12,
          ),
          itemCount: categoriesList.length,
          itemBuilder: (context, index) {
            final cat = categoriesList[index];
            final rawName = cat['rawName'] as String;

            return GestureDetector(
              onTap: () => context.push('/category/products?name=${Uri.encodeComponent(rawName)}'),
              child: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: cat['bgColor'] as Color,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: cat['borderColor'] as Color, width: 1.2),
                  boxShadow: const [
                    BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2)),
                  ],
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Expanded(
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Image.network(
                          cat['image_url'] as String,
                          fit: BoxFit.contain,
                          errorBuilder: (_, __, ___) => const Icon(Icons.shopping_bag_outlined, color: Colors.grey, size: 28),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      cat['name'] as String,
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w800, color: Color(0xFF1F2937), height: 1.2),
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
    );
  }
}

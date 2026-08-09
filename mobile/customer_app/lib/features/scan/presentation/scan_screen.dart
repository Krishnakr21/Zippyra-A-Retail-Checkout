import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';

import '../../cart/presentation/bloc/cart_bloc.dart';

class ScanScreen extends StatefulWidget {
  final String? targetProductName;
  final String? targetBarcode;
  final String? targetPrice;
  final String? targetImageUrl;

  const ScanScreen({
    super.key,
    this.targetProductName,
    this.targetBarcode,
    this.targetPrice,
    this.targetImageUrl,
  });

  @override
  State<ScanScreen> createState() => _ScanScreenState();
}

class _ScanScreenState extends State<ScanScreen> with SingleTickerProviderStateMixin {
  final MobileScannerController _scannerController = MobileScannerController(
    detectionSpeed: DetectionSpeed.normal,
    facing: CameraFacing.back,
  );

  bool _isProcessing = false;
  bool _showSuccessOverlay = false;
  Map<String, String>? _lastScannedProduct;

  final Map<String, Map<String, String>> _knownBarcodes = {
    '8904335601951': {
      'name': '100% Rolled Oats (Yoga Bar)',
      'price': '₹493',
      'barcode': '8904335601951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/433/560/1951/front_en.14.400.jpg',
    },
    '8901063325036': {
      'name': 'Britannia Toastea Bake Rusk 250g',
      'price': '₹212',
      'barcode': '8901063325036',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/332/5036/front_en.10.400.jpg',
    },
    '8906136651951': {
      'name': 'Pintola High Protein Oats 400g',
      'price': '₹483',
      'barcode': '8906136651951',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/613/665/1951/front_en.10.400.jpg',
    },
    '8901262010054': {
      'name': 'Amul Taaza Toned Milk 1L',
      'price': '₹68',
      'barcode': '8901262010054',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/126/201/0054/front_en.10.400.jpg',
    },
    '8901052000155': {
      'name': 'Tata Tea Gold Leaf 500g',
      'price': '₹315',
      'barcode': '8901052000155',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/105/200/0155/front_en.10.400.jpg',
    },
    '8901063012011': {
      'name': 'Britannia Good Day Cashew 200g',
      'price': '₹45',
      'barcode': '8901063012011',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/106/301/2011/front_en.10.400.jpg',
    },
    '8906007280015': {
      'name': 'Fortune Sunlite Sunflower Oil 1L',
      'price': '₹142',
      'barcode': '8906007280015',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/600/728/0015/front_en.10.400.jpg',
    },
    '8901088034593': {
      'name': 'Saffola Active Edible Oil 1L',
      'price': '₹356',
      'barcode': '8901088034593',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/108/803/4593/front_en.10.400.jpg',
    },
    '8901030678912': {
      'name': 'Amul Pasteurised Butter 500g',
      'price': '₹280',
      'barcode': '8901030678912',
      'image_url': 'https://images.openfoodfacts.org/images/products/890/103/067/8912/front_en.10.400.jpg',
    },
  };

  @override
  void initState() {
    super.initState();
    // If target product passed from product card ADD button, pre-populate last scanned item
    if ((widget.targetProductName ?? '').isNotEmpty) {
      _lastScannedProduct = {
        'name': widget.targetProductName!,
        'price': widget.targetPrice ?? '₹493',
        'barcode': widget.targetBarcode ?? '8904335601951',
        'image_url': widget.targetImageUrl ?? '',
      };
    }
  }

  @override
  void dispose() {
    _scannerController.dispose();
    super.dispose();
  }

  void _onProcessBarcode(String rawCode) {
    if (_isProcessing) return;
    final code = rawCode.trim();
    if (code.isEmpty) return;

    final matched = _knownBarcodes[code] ?? {
      'name': widget.targetProductName ?? 'Scanned Store Product',
      'price': widget.targetPrice ?? '₹250',
      'barcode': code,
      'image_url': widget.targetImageUrl ?? '',
    };

    setState(() {
      _isProcessing = true;
      _lastScannedProduct = matched;
      _showSuccessOverlay = true;
    });

    // Dispatch to CartBloc
    context.read<CartBloc>().add(
          ItemScanned(
            storeId: 'store-1',
            barcode: code,
            qty: 1,
          ),
        );

    // Hide success pulse overlay after 1.2s
    Future.delayed(const Duration(milliseconds: 1200), () {
      if (mounted) {
        setState(() {
          _isProcessing = false;
          _showSuccessOverlay = false;
        });
      }
    });
  }

  void _showSkuPickerSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xFF1E293B),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) {
        return Container(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Select Product SKU to Scan',
                    style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.white54),
                    onPressed: () => Navigator.pop(ctx),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              SizedBox(
                height: 240,
                child: ListView(
                  children: _knownBarcodes.values.map((item) {
                    return Container(
                      margin: const EdgeInsets.only(bottom: 8),
                      decoration: BoxDecoration(
                        color: const Color(0xFF334155),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: ListTile(
                        leading: Container(
                          width: 40,
                          height: 40,
                          padding: const EdgeInsets.all(4),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: (item['image_url'] ?? '').isNotEmpty
                              ? Image.network(item['image_url']!, fit: BoxFit.contain)
                              : const Icon(Icons.inventory_2_outlined, color: Colors.grey),
                        ),
                        title: Text(item['name']!, style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.bold)),
                        subtitle: Text('SKU: ${item['barcode']} · ${item['price']}', style: const TextStyle(color: Colors.white70, fontSize: 11)),
                        trailing: const Icon(Icons.qr_code_scanner, color: Color(0xFF16A34A)),
                        onTap: () {
                          Navigator.pop(ctx);
                          _onProcessBarcode(item['barcode']!);
                        },
                      ),
                    );
                  }).toList(),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          children: [
            // 1. Live Camera Viewfinder Layer
            Positioned.fill(
              child: MobileScanner(
                controller: _scannerController,
                onDetect: (capture) {
                  final List<Barcode> barcodes = capture.barcodes;
                  for (final b in barcodes) {
                    final code = b.rawValue;
                    if (code != null && code.isNotEmpty && !_isProcessing) {
                      _onProcessBarcode(code);
                      break;
                    }
                  }
                },
                errorBuilder: (context, error, child) {
                  return Container(
                    color: const Color(0xFF0F172A),
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.camera_alt_outlined, color: Colors.white38, size: 64),
                          const SizedBox(height: 12),
                          const Text(
                            'Scan & Go Camera Feed Active',
                            style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16),
                          ),
                          const SizedBox(height: 6),
                          Text(
                            'Point camera at barcode or tap "Enter SKU"',
                            style: TextStyle(color: Colors.white.withOpacity(0.6), fontSize: 12),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),

            // 2. Center Reticle & Scanning Frame (Figma Screen 8.1)
            Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Container(
                    width: 260,
                    height: 260,
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: _showSuccessOverlay ? const Color(0xFF16A34A) : const Color(0xFF38BDF8),
                        width: 3.5,
                      ),
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                          color: (_showSuccessOverlay ? const Color(0xFF16A34A) : const Color(0xFF0284C7)).withOpacity(0.35),
                          blurRadius: 24,
                          spreadRadius: 4,
                        ),
                      ],
                    ),
                    child: Center(
                      child: Container(
                        width: 230,
                        height: 2,
                        color: (_showSuccessOverlay ? const Color(0xFF16A34A) : const Color(0xFF38BDF8)).withOpacity(0.9),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    decoration: BoxDecoration(
                      color: Colors.black.withOpacity(0.75),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      _showSuccessOverlay ? 'Item Scanned ✓' : 'Point at barcode to scan',
                      style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w600),
                    ),
                  ),
                ],
              ),
            ),

            // 3. Screen 8.2 SCAN SUCCESS Green Overlay Pulse
            if (_showSuccessOverlay)
              Positioned.fill(
                child: Container(
                  color: const Color(0xFF16A34A).withOpacity(0.85),
                  child: Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Container(
                          width: 84,
                          height: 84,
                          decoration: const BoxDecoration(
                            color: Colors.white,
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(Icons.check_rounded, color: Color(0xFF16A34A), size: 54),
                        ),
                        const SizedBox(height: 20),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.85),
                            borderRadius: BorderRadius.circular(30),
                          ),
                          child: Column(
                            children: [
                              Row(
                                mainAxisSize: MainAxisSize.min,
                                children: const [
                                  Text('Item Added!', style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w900)),
                                  SizedBox(width: 6),
                                  Icon(Icons.check_circle, color: Color(0xFF22C55E), size: 18),
                                ],
                              ),
                              if (_lastScannedProduct != null) ...[
                                const SizedBox(height: 4),
                                Text(
                                  '${_lastScannedProduct!['name']} · ${_lastScannedProduct!['price']}',
                                  style: const TextStyle(color: Colors.white70, fontSize: 12),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),

            // 4. Top Header Navigation (Figma Screen 8.1 Header)
            Positioned(
              top: 12,
              left: 16,
              right: 16,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  GestureDetector(
                    onTap: () => context.pop(),
                    child: Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: Colors.black.withOpacity(0.65),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.arrow_back, color: Colors.white, size: 20),
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.black.withOpacity(0.65),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: Colors.white24),
                    ),
                    child: Row(
                      children: const [
                        Icon(Icons.qr_code_scanner, color: Color(0xFF38BDF8), size: 16),
                        SizedBox(width: 6),
                        Text(
                          'Scan & Go Mode',
                          style: TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                  ),
                  Row(
                    children: [
                      IconButton(
                        icon: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.65),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(Icons.tune_rounded, color: Colors.white, size: 18),
                        ),
                        onPressed: _showSkuPickerSheet,
                      ),
                      IconButton(
                        icon: Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.65),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(Icons.lightbulb_outline, color: Colors.white, size: 18),
                        ),
                        onPressed: () => _scannerController.toggleTorch(),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // 5. Scanned Product Pop-Up Card & Bottom Bar (Figma Screen 8.1 Bottom Engine)
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Pop-Up Card above confirm button for scanned product
                  if (_lastScannedProduct != null)
                    Container(
                      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFF1E293B),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: const Color(0xFF334155)),
                        boxShadow: const [
                          BoxShadow(color: Colors.black45, blurRadius: 10, offset: Offset(0, -2)),
                        ],
                      ),
                      child: Row(
                        children: [
                          Container(
                            width: 44,
                            height: 44,
                            padding: const EdgeInsets.all(4),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: (_lastScannedProduct!['image_url'] ?? '').isNotEmpty
                                ? Image.network(_lastScannedProduct!['image_url']!, fit: BoxFit.contain)
                                : const Icon(Icons.inventory_2_outlined, color: Colors.grey),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  _lastScannedProduct!['name']!,
                                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 13),
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                const SizedBox(height: 2),
                                Row(
                                  children: [
                                    const Text('Just scanned', style: TextStyle(color: Color(0xFF86EFAC), fontSize: 11, fontWeight: FontWeight.w600)),
                                    const SizedBox(width: 4),
                                    const Icon(Icons.check, color: Color(0xFF86EFAC), size: 12),
                                    const Spacer(),
                                    Text(_lastScannedProduct!['price']!, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 13)),
                                  ],
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),

                  // Bottom Action Summary Bar (Screen 8.1)
                  BlocBuilder<CartBloc, CartState>(
                    builder: (context, state) {
                      int itemCount = 0;
                      int totalPaise = 0;

                      if (state is CartLoaded) {
                        itemCount = state.summary.itemCount;
                        totalPaise = state.summary.totalPaise;
                      } else if (state is CartCouponError) {
                        itemCount = state.summary?.itemCount ?? 0;
                        totalPaise = state.summary?.totalPaise ?? 0;
                      }

                      final totalRs = (totalPaise / 100).round();

                      return Container(
                        padding: const EdgeInsets.all(16),
                        decoration: const BoxDecoration(
                          color: Color(0xFF0F172A),
                          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
                          border: Border(top: BorderSide(color: Color(0xFF1E293B))),
                        ),
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '$itemCount ITEMS IN CART',
                                      style: const TextStyle(color: Colors.white60, fontSize: 10, fontWeight: FontWeight.w800, letterSpacing: 0.5),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      '₹$totalRs',
                                      style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w900),
                                    ),
                                  ],
                                ),
                                Row(
                                  children: [
                                    OutlinedButton.icon(
                                      onPressed: _showSkuPickerSheet,
                                      icon: const Icon(Icons.qr_code, size: 16, color: Colors.white),
                                      label: const Text('Enter SKU', style: TextStyle(color: Colors.white, fontSize: 11)),
                                      style: OutlinedButton.styleFrom(
                                        side: const BorderSide(color: Colors.white30),
                                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    ElevatedButton(
                                      onPressed: () => context.go('/cart'),
                                      style: ElevatedButton.styleFrom(
                                        backgroundColor: const Color(0xFF16A34A),
                                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                                      ),
                                      child: Row(
                                        children: const [
                                          Text('View Cart', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 12)),
                                          SizedBox(width: 4),
                                          Icon(Icons.arrow_forward_rounded, color: Colors.white, size: 16),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

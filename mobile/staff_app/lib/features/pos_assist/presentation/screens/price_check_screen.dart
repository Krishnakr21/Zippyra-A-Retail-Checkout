import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../bloc/price_check_bloc.dart';
import '../bloc/price_check_event.dart';
import '../bloc/price_check_state.dart';

class PriceCheckScreen extends StatefulWidget {
  final String storeId;

  const PriceCheckScreen({Key? key, this.storeId = 'store-demo-1'}) : super(key: key);

  @override
  State<PriceCheckScreen> createState() => _PriceCheckScreenState();
}

class _PriceCheckScreenState extends State<PriceCheckScreen> {
  final TextEditingController _manualController = TextEditingController();
  final MobileScannerController _scannerController = MobileScannerController();
  bool _showScanner = true;

  @override
  void dispose() {
    _manualController.dispose();
    _scannerController.dispose();
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    final List<Barcode> barcodes = capture.barcodes;
    for (final barcode in barcodes) {
      if (barcode.rawValue != null && barcode.rawValue!.isNotEmpty) {
        context.read<PriceCheckBloc>().add(
              BarcodeScanned(storeId: widget.storeId, barcode: barcode.rawValue!),
            );
        break;
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Price Check'),
        actions: [
          IconButton(
            icon: Icon(_showScanner ? Icons.keyboard : Icons.camera_alt),
            onPressed: () {
              setState(() {
                _showScanner = !_showScanner;
              });
            },
          ),
        ],
      ),
      body: Column(
        children: [
          if (_showScanner)
            SizedBox(
              height: 220,
              child: MobileScanner(
                controller: _scannerController,
                onDetect: _onDetect,
              ),
            ),
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _manualController,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      hintText: 'Enter barcode manually...',
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                ElevatedButton(
                  key: const Key('btn_manual_lookup'),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
                    backgroundColor: Colors.indigo,
                  ),
                  onPressed: () {
                    context.read<PriceCheckBloc>().add(
                          ManualBarcodeSubmitted(
                            storeId: widget.storeId,
                            barcode: _manualController.text,
                          ),
                        );
                  },
                  child: const Text('Check', style: TextStyle(color: Colors.white)),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: BlocBuilder<PriceCheckBloc, PriceCheckState>(
              builder: (context, state) {
                if (state is PriceCheckLoading) {
                  return const Center(child: CircularProgressIndicator());
                }

                if (state is PriceCheckNotFound) {
                  return Center(
                    child: Padding(
                      padding: const EdgeInsets.all(24.0),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.search_off, size: 64, color: Colors.amber),
                          const SizedBox(height: 16),
                          Text(
                            'Product not found for "${state.barcode}".',
                            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 8),
                          const Text(
                            'Try scanning again or check with your manager.',
                            style: TextStyle(color: Colors.grey, fontSize: 14),
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),
                  );
                }

                if (state is PriceCheckFailed) {
                  return Center(
                    child: Text(
                      state.message,
                      style: const TextStyle(color: Colors.red, fontSize: 16),
                    ),
                  );
                }

                if (state is PriceCheckFound) {
                  final p = state.product;
                  final priceStr = CurrencyFormatter.formatPaise(p.pricePaise);
                  final mrpStr = CurrencyFormatter.formatPaise(p.mrpPaise);

                  return SingleChildScrollView(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (state.fetchedFromRemote)
                          Container(
                            padding: const EdgeInsets.all(8),
                            margin: const EdgeInsets.only(bottom: 16),
                            decoration: BoxDecoration(
                              color: Colors.amber.shade50,
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(color: Colors.amber),
                            ),
                            child: const Row(
                              children: [
                                Icon(Icons.cloud_download, color: Colors.amber),
                                SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    'Fetched from remote catalog (Local sync pending)',
                                    style: TextStyle(fontSize: 12, color: Colors.black87),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        Text(
                          p.name,
                          style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Barcode: ${p.barcode}',
                          style: const TextStyle(fontSize: 14, color: Colors.grey),
                        ),
                        const SizedBox(height: 24),
                        Row(
                          children: [
                            Text(
                              priceStr,
                              style: const TextStyle(
                                fontSize: 36,
                                fontWeight: FontWeight.w800,
                                color: Colors.green,
                              ),
                            ),
                            const SizedBox(width: 16),
                            if (p.mrpPaise > p.pricePaise)
                              Text(
                                mrpStr,
                                style: const TextStyle(
                                  fontSize: 22,
                                  color: Colors.grey,
                                  decoration: TextDecoration.lineThrough,
                                ),
                              ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        ListTile(
                          contentPadding: EdgeInsets.zero,
                          title: const Text('HSN / GST Rate'),
                          subtitle: Text('${p.hsnCode} / ${p.gstRatePercent}%'),
                        ),
                      ],
                    ),
                  );
                }

                return const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.barcode_reader, size: 64, color: Colors.grey),
                      SizedBox(height: 16),
                      Text(
                        'Scan a barcode or enter manually to check price',
                        style: TextStyle(fontSize: 16, color: Colors.grey),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

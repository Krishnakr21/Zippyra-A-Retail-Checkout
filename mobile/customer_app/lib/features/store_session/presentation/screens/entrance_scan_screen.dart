import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:permission_handler/permission_handler.dart';
import '../bloc/store_session_bloc.dart';

class EntranceScanScreen extends StatefulWidget {
  const EntranceScanScreen({super.key});

  @override
  State<EntranceScanScreen> createState() => _EntranceScanScreenState();
}

class _EntranceScanScreenState extends State<EntranceScanScreen> {
  bool _hasPermission = false;
  bool _isScanning = true;
  final MobileScannerController _scannerController = MobileScannerController();

  @override
  void initState() {
    super.initState();
    _checkPermissions();
  }

  @override
  void dispose() {
    _scannerController.dispose();
    super.dispose();
  }

  Future<void> _checkPermissions() async {
    final cameraStatus = await Permission.camera.status;
    final locationStatus = await Permission.locationWhenInUse.status;

    if (cameraStatus.isGranted && locationStatus.isGranted) {
      setState(() => _hasPermission = true);
    } else {
      final cameraReq = await Permission.camera.request();
      final locationReq = await Permission.locationWhenInUse.request();
      setState(() {
        _hasPermission = cameraReq.isGranted && locationReq.isGranted;
      });
    }
  }

  void _onDetect(BarcodeCapture capture) {
    if (!_isScanning) return;
    final List<Barcode> barcodes = capture.barcodes;
    for (final barcode in barcodes) {
      if (barcode.rawValue != null && barcode.rawValue!.isNotEmpty) {
        setState(() => _isScanning = false);
        final token = barcode.rawValue!;
        context.read<StoreSessionBloc>().add(EntranceQrScanned(token));
        context.pushReplacement('/store/binding');
        break;
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Scan Entrance QR')),
      body: !_hasPermission
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.security, size: 64, color: Colors.orange),
                    const SizedBox(height: 16),
                    const Text(
                      'Camera & Location Permissions Required',
                      style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    const Text(
                      'Please enable Camera and Location permissions in Settings to scan entrance QR codes and verify store entrance.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey),
                    ),
                    const SizedBox(height: 24),
                    ElevatedButton(
                      onPressed: openAppSettings,
                      child: const Text('Open Settings'),
                    ),
                  ],
                ),
              ),
            )
          : Stack(
              children: [
                MobileScanner(
                  controller: _scannerController,
                  onDetect: _onDetect,
                ),
                Center(
                  child: Container(
                    width: 250,
                    height: 250,
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.blue, width: 4),
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
                const Positioned(
                  bottom: 40,
                  left: 0,
                  right: 0,
                  child: Text(
                    'Align Entrance QR Code within frame',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.white, fontSize: 16, backgroundColor: Colors.black54),
                  ),
                ),
              ],
            ),
    );
  }
}

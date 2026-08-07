import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:go_router/go_router.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

class PermissionsScreen extends StatefulWidget {
  const PermissionsScreen({super.key});

  @override
  State<PermissionsScreen> createState() => _PermissionsScreenState();
}

class _PermissionsScreenState extends State<PermissionsScreen> {
  bool _locationGranted = false;
  bool _cameraGranted = false;
  bool _isChecking = true;

  @override
  void initState() {
    super.initState();
    _checkRealSystemPermissions();
  }

  Future<void> _checkRealSystemPermissions() async {
    setState(() => _isChecking = true);
    bool locOk = false;
    bool camOk = false;

    try {
      final locPerm = await Geolocator.checkPermission();
      locOk = (locPerm == LocationPermission.always || locPerm == LocationPermission.whileInUse);
    } catch (_) {}

    try {
      final camStatus = await Permission.camera.status;
      camOk = camStatus.isGranted;
    } catch (_) {}

    if (mounted) {
      setState(() {
        _locationGranted = locOk;
        _cameraGranted = camOk;
        _isChecking = false;
      });
    }
  }

  Future<void> _requestLocationPermission() async {
    try {
      LocationPermission locPerm = await Geolocator.checkPermission();
      if (locPerm == LocationPermission.denied) {
        locPerm = await Geolocator.requestPermission();
      }
      if (locPerm == LocationPermission.deniedForever) {
        await openAppSettings();
      }
      final isOk = (locPerm == LocationPermission.always || locPerm == LocationPermission.whileInUse);
      setState(() => _locationGranted = isOk);
      if (!isOk) {
        _showPermissionDeniedBanner('Location');
      }
    } catch (e) {
      final status = await Permission.location.request();
      setState(() => _locationGranted = status.isGranted);
    }
  }

  Future<void> _requestCameraPermission() async {
    try {
      var camStatus = await Permission.camera.status;
      if (!camStatus.isGranted) {
        camStatus = await Permission.camera.request();
      }
      if (camStatus.isPermanentlyDenied) {
        await openAppSettings();
      }
      setState(() => _cameraGranted = camStatus.isGranted);
      if (!camStatus.isGranted) {
        _showPermissionDeniedBanner('Camera');
      }
    } catch (e) {
      final status = await Permission.camera.request();
      setState(() => _cameraGranted = status.isGranted);
    }
  }

  void _showPermissionDeniedBanner(String permissionName) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('⚠️ System $permissionName permission required. Please grant in browser/system settings.'),
        backgroundColor: ZippyraColors.errorRed,
        behavior: SnackBarBehavior.floating,
        action: SnackBarAction(
          label: 'SETTINGS',
          textColor: Colors.white,
          onPressed: () => openAppSettings(),
        ),
      ),
    );
  }

  void _onVerifyAndProceed() async {
    await _checkRealSystemPermissions();
    if (!_locationGranted || !_cameraGranted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('🛑 Access Blocked: System Location and Camera permissions MUST be granted in system settings.'),
          backgroundColor: ZippyraColors.errorRed,
          behavior: SnackBarBehavior.floating,
          duration: Duration(seconds: 4),
        ),
      );
      return;
    }
    context.go('/home');
  }

  @override
  Widget build(BuildContext context) {
    final bool allPermissionsGranted = _locationGranted && _cameraGranted;

    return Scaffold(
      backgroundColor: const Color(0xFF0F172A),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 20.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 20),
              // Security Icon
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: ZippyraColors.primaryBlue.withOpacity(0.15),
                  shape: BoxShape.circle,
                  border: Border.all(color: ZippyraColors.primaryBlue.withOpacity(0.3)),
                ),
                child: const Icon(
                  Icons.security_rounded,
                  size: 36,
                  color: ZippyraColors.primaryBlue,
                ),
              ),
              const SizedBox(height: 20),
              const Text(
                'System Permissions Required',
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.w800,
                  color: Colors.white,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Zippyra forcefully enforces system Location & Camera permissions for in-store geofencing, BLE store binding, and instant barcode checkout.',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.white.withOpacity(0.7),
                  height: 1.4,
                ),
              ),
              const SizedBox(height: 32),

              // Location Permission System Card
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.06),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: _locationGranted ? ZippyraColors.successGreen : ZippyraColors.primaryBlue.withOpacity(0.4),
                    width: 1.5,
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: ZippyraColors.primaryBlue.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Icon(
                        _locationGranted ? Icons.check_circle_rounded : Icons.location_on_rounded,
                        color: _locationGranted ? ZippyraColors.successGreen : ZippyraColors.primaryBlue,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _locationGranted ? 'Location Access Granted ✅' : 'Location Access (System Mandatory)',
                            style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Triggers system prompt for BLE & Geofence store binding',
                            style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.6)),
                          ),
                        ],
                      ),
                    ),
                    ElevatedButton(
                      onPressed: _requestLocationPermission,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _locationGranted ? ZippyraColors.successGreen.withOpacity(0.2) : ZippyraColors.primaryBlue,
                        foregroundColor: _locationGranted ? ZippyraColors.successGreen : Colors.white,
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      child: Text(_locationGranted ? 'Granted' : 'Grant OS'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // Camera Permission System Card
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.06),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: _cameraGranted ? ZippyraColors.successGreen : ZippyraColors.accentOrange.withOpacity(0.4),
                    width: 1.5,
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: ZippyraColors.accentOrange.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Icon(
                        _cameraGranted ? Icons.check_circle_rounded : Icons.camera_alt_rounded,
                        color: _cameraGranted ? ZippyraColors.successGreen : ZippyraColors.accentOrange,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            _cameraGranted ? 'Camera Access Granted ✅' : 'Camera Access (System Mandatory)',
                            style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Triggers system prompt for instant barcode scanning',
                            style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.6)),
                          ),
                        ],
                      ),
                    ),
                    ElevatedButton(
                      onPressed: _requestCameraPermission,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _cameraGranted ? ZippyraColors.successGreen.withOpacity(0.2) : ZippyraColors.accentOrange,
                        foregroundColor: _cameraGranted ? ZippyraColors.successGreen : Colors.white,
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      child: Text(_cameraGranted ? 'Granted' : 'Grant OS'),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 16),
              Center(
                child: TextButton.icon(
                  onPressed: () => openAppSettings(),
                  icon: const Icon(Icons.settings_suggest_rounded, color: Colors.white70, size: 18),
                  label: const Text('Open System App Settings ⚙️', style: TextStyle(color: Colors.white70, fontSize: 13)),
                ),
              ),

              const Spacer(),

              // Forceful Proceed Button
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton(
                  onPressed: _onVerifyAndProceed,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: allPermissionsGranted ? ZippyraColors.primaryBlue : const Color(0xFF1E293B),
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(14),
                    ),
                    elevation: allPermissionsGranted ? 4 : 0,
                  ),
                  child: Text(
                    allPermissionsGranted ? 'Verify & Continue to Home 🚀' : 'Grant System Permissions to Proceed 🔒',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: allPermissionsGranted ? Colors.white : Colors.white38,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }
}

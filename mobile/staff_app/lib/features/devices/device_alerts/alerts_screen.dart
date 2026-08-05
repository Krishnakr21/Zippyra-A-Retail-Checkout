import 'package:flutter/material.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zcard.dart';
import 'package:zippyra_core/widgets/zbadge.dart';

class DeviceAlertsScreen extends StatelessWidget {
  const DeviceAlertsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('IoT Device Alerts (MQTT)')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: const [
          ZCard(
            child: ListTile(
              leading: Icon(Icons.sensor_door, color: ZippyraColors.errorRed, size: 36),
              title: Text('Turnstile Gate #02 Offline', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: Text('MQTT Heartbeat Lost • 2m ago'),
              trailing: ZBadge(text: 'CRITICAL', color: ZippyraColors.errorRed),
            ),
          ),
        ],
      ),
    );
  }
}

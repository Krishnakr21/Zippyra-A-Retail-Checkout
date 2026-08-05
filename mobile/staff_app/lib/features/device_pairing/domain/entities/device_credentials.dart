class DeviceCredentials {
  final String deviceId;
  final String deviceJwt;
  final String certPem;
  final String privateKeyPem;
  final String rootCaPem;
  final String mqttEndpoint;

  const DeviceCredentials({
    required this.deviceId,
    required this.deviceJwt,
    required this.certPem,
    required this.privateKeyPem,
    required this.rootCaPem,
    required this.mqttEndpoint,
  });

  factory DeviceCredentials.fromJson(Map<String, dynamic> json) {
    return DeviceCredentials(
      deviceId: json['device_id'] as String? ?? '',
      deviceJwt: json['device_jwt'] as String? ?? '',
      certPem: json['cert_pem'] as String? ?? '',
      privateKeyPem: json['private_key_pem'] as String? ?? '',
      rootCaPem: json['root_ca_pem'] as String? ?? '',
      mqttEndpoint: json['mqtt_endpoint'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'device_id': deviceId,
        'device_jwt': deviceJwt,
        'cert_pem': certPem,
        'private_key_pem': privateKeyPem,
        'root_ca_pem': rootCaPem,
        'mqtt_endpoint': mqttEndpoint,
      };
}

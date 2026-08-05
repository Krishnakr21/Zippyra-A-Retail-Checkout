import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class AppConfig {
  static const String _appEnvDefine = String.fromEnvironment('APP_ENV', defaultValue: '');
  static const String _appVersionNameDefine = String.fromEnvironment('APP_VERSION_NAME', defaultValue: '');
  static const String _baseUrlDevDefine = String.fromEnvironment('API_BASE_URL_DEV', defaultValue: '');
  static const String _baseUrlStagingDefine = String.fromEnvironment('API_BASE_URL_STAGING', defaultValue: '');
  static const String _baseUrlProdDefine = String.fromEnvironment('API_BASE_URL_PROD', defaultValue: '');
  static const String _googleOAuthServerClientIdDefine = String.fromEnvironment('GOOGLE_OAUTH_SERVER_CLIENT_ID', defaultValue: '');
  static const String _googleClientIdDefine = String.fromEnvironment('GOOGLE_OAUTH_CLIENT_ID', defaultValue: '');
  static const String _razorpayKeyIdDefine = String.fromEnvironment('RAZORPAY_KEY_ID', defaultValue: '');
  static const String _certPinPrimaryDefine = String.fromEnvironment('CERT_PIN_SHA256_PRIMARY', defaultValue: '');
  static const String _certPinBackupDefine = String.fromEnvironment('CERT_PIN_SHA256_BACKUP', defaultValue: '');

  static const String _mqttBrokerUrlDefine = String.fromEnvironment('MQTT_BROKER_URL', defaultValue: '');
  static const String _sentryDsnMobileDefine = String.fromEnvironment('SENTRY_DSN_MOBILE', defaultValue: '');
  static const String _googleMapsApiKeyAndroidDefine = String.fromEnvironment('GOOGLE_MAPS_API_KEY_ANDROID', defaultValue: '');
  static const String _googleMapsApiKeyIosDefine = String.fromEnvironment('GOOGLE_MAPS_API_KEY_IOS', defaultValue: '');

  static Future<void> initialize() async {
    if (kDebugMode) {
      try {
        await dotenv.load(fileName: '.env');
      } catch (_) {
        // .env file missing in debug mode -> proceed with defaults / dart-defines
      }
    }
  }

  static String get appEnv {
    if (_appEnvDefine.isNotEmpty) return _appEnvDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['APP_ENV'] ?? 'development';
    }
    return 'development';
  }

  static String get appVersionName {
    if (_appVersionNameDefine.isNotEmpty) return _appVersionNameDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['APP_VERSION_NAME'] ?? '1.0.0';
    }
    return '1.0.0';
  }

  static String get appMinSupportedVersion {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['APP_MIN_SUPPORTED_VERSION'] ?? '1.0.0';
    }
    return '1.0.0';
  }

  static String get baseUrl {
    final env = appEnv;
    if (env == 'production') {
      return apiBaseUrlProd;
    } else if (env == 'staging') {
      return apiBaseUrlStaging;
    }
    return apiBaseUrlDev;
  }

  static String get apiBaseUrlDev {
    if (_baseUrlDevDefine.isNotEmpty) {
      if (kIsWeb && (_baseUrlDevDefine.contains('10.0.2.2') || _baseUrlDevDefine.contains('192.168.'))) {
        return _baseUrlDevDefine.replaceAll('10.0.2.2', 'localhost').replaceAll(RegExp(r'192\.168\.\d+\.\d+'), 'localhost');
      }
      return _baseUrlDevDefine;
    }
    if (kDebugMode && dotenv.isInitialized) {
      final envUrl = dotenv.env['API_BASE_URL_DEV'];
      if (envUrl != null && envUrl.isNotEmpty) {
        if (kIsWeb && (envUrl.contains('10.0.2.2') || envUrl.contains('192.168.'))) {
          return envUrl.replaceAll('10.0.2.2', 'localhost').replaceAll(RegExp(r'192\.168\.\d+\.\d+'), 'localhost');
        }
        return envUrl;
      }
    }
    return 'http://localhost:8080';
  }

  static String get apiBaseUrlStaging {
    if (_baseUrlStagingDefine.isNotEmpty) return _baseUrlStagingDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['API_BASE_URL_STAGING'] ?? 'https://staging-api.zippyra.com';
    }
    return 'https://staging-api.zippyra.com';
  }

  static String get apiBaseUrlProd {
    if (_baseUrlProdDefine.isNotEmpty) return _baseUrlProdDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['API_BASE_URL_PROD'] ?? 'https://api.zippyra.com';
    }
    return 'https://api.zippyra.com';
  }

  static String get wsBaseUrlProd {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['WS_BASE_URL_PROD'] ?? 'wss://api.zippyra.com/ws';
    }
    return 'wss://api.zippyra.com/ws';
  }

  static String get certPinSha256Primary {
    if (_certPinPrimaryDefine.isNotEmpty) return _certPinPrimaryDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['CERT_PIN_SHA256_PRIMARY'] ?? '';
    }
    return '';
  }

  static String get certPinSha256Backup {
    if (_certPinBackupDefine.isNotEmpty) return _certPinBackupDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['CERT_PIN_SHA256_BACKUP'] ?? '';
    }
    return '';
  }

  static String get googleOAuthServerClientId {
    if (_googleOAuthServerClientIdDefine.isNotEmpty) return _googleOAuthServerClientIdDefine;
    if (_googleClientIdDefine.isNotEmpty) return _googleClientIdDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['GOOGLE_OAUTH_SERVER_CLIENT_ID'] ?? dotenv.env['GOOGLE_OAUTH_CLIENT_ID'] ?? '';
    }
    return '';
  }

  static String get googleOAuthIosClientId {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['GOOGLE_OAUTH_IOS_CLIENT_ID'] ?? '';
    }
    return '';
  }

  static String get googleOAuthAndroidClientId {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['GOOGLE_OAUTH_ANDROID_CLIENT_ID'] ?? '';
    }
    return '';
  }

  static String get razorpayKeyId {
    if (_razorpayKeyIdDefine.isNotEmpty) return _razorpayKeyIdDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['RAZORPAY_KEY_ID'] ?? '';
    }
    return '';
  }

  static String get googleMapsApiKeyAndroid {
    if (_googleMapsApiKeyAndroidDefine.isNotEmpty) return _googleMapsApiKeyAndroidDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['GOOGLE_MAPS_API_KEY_ANDROID'] ?? '';
    }
    return '';
  }

  static String get googleMapsApiKeyIos {
    if (_googleMapsApiKeyIosDefine.isNotEmpty) return _googleMapsApiKeyIosDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['GOOGLE_MAPS_API_KEY_IOS'] ?? '';
    }
    return '';
  }

  static String get fcmVapidKey {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['FCM_VAPID_KEY'] ?? '';
    }
    return '';
  }

  static String get playIntegrityCloudProjectNumber {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['PLAY_INTEGRITY_CLOUD_PROJECT_NUMBER'] ?? '';
    }
    return '';
  }

  static bool get rootJailbreakCheckEnabled {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['ROOT_JAILBREAK_CHECK_ENABLED'];
      if (val != null) return val.toLowerCase() == 'true';
    }
    return true;
  }

  static bool get localCatalogSyncOnShiftStart {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['LOCAL_CATALOG_SYNC_ON_SHIFT_START'];
      if (val != null) return val.toLowerCase() == 'true';
    }
    return true;
  }

  static int get offlineQueueRetryIntervalSeconds {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['OFFLINE_QUEUE_RETRY_INTERVAL_SECONDS'];
      if (val != null) return int.tryParse(val) ?? 15;
    }
    return 15;
  }

  static int get offlineQueueMaxRetryAttempts {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['OFFLINE_QUEUE_MAX_RETRY_ATTEMPTS'];
      if (val != null) return int.tryParse(val) ?? 10;
    }
    return 10;
  }

  static String get mqttBrokerUrl {
    if (_mqttBrokerUrlDefine.isNotEmpty) return _mqttBrokerUrlDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['MQTT_BROKER_URL'] ?? '';
    }
    return '';
  }

  static String get mqttClientCertPath {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['MQTT_CLIENT_CERT_PATH'] ?? 'assets/certs/staff-device-cert.pem';
    }
    return 'assets/certs/staff-device-cert.pem';
  }

  static String get mqttClientKeyPath {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['MQTT_CLIENT_KEY_PATH'] ?? 'assets/certs/staff-device-key.pem';
    }
    return 'assets/certs/staff-device-key.pem';
  }

  static String get mqttTopicPrefix {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['MQTT_TOPIC_PREFIX'] ?? 'zippyra';
    }
    return 'zippyra';
  }

  static int get staffSessionIdleTimeoutMinutes {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['STAFF_SESSION_IDLE_TIMEOUT_MINUTES'];
      if (val != null) return int.tryParse(val) ?? 30;
    }
    return 30;
  }

  static String get sentryDsnMobile {
    if (_sentryDsnMobileDefine.isNotEmpty) return _sentryDsnMobileDefine;
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['SENTRY_DSN_MOBILE'] ?? '';
    }
    return '';
  }

  static double get sentryTracesSampleRate {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['SENTRY_TRACES_SAMPLE_RATE'];
      if (val != null) return double.tryParse(val) ?? 0.2;
    }
    return 0.2;
  }

  static String get defaultLocale {
    if (kDebugMode && dotenv.isInitialized) {
      return dotenv.env['DEFAULT_LOCALE'] ?? 'en';
    }
    return 'en';
  }

  static int get offlineSyncJitterMaxSeconds {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['OFFLINE_SYNC_JITTER_MAX_SECONDS'];
      if (val != null) return int.tryParse(val) ?? 30;
    }
    return 30;
  }

  static int get offerCacheMaxAgeSeconds {
    if (kDebugMode && dotenv.isInitialized) {
      final val = dotenv.env['OFFER_CACHE_MAX_AGE_SECONDS'];
      if (val != null) return int.tryParse(val) ?? 300;
    }
    return 300;
  }

  /// Startup assertion called before runApp()
  static void validate() {
    final env = appEnv;
    if (env == 'production') {
      if (certPinSha256Primary.isEmpty) {
        throw StateError(
          'CRITICAL SECURITY ERROR: Production mobile build running without CERT_PIN_SHA256_PRIMARY configured! '
          'Release builds must be compiled with --dart-define-from-file=env/prod.json containing certificate pins.',
        );
      }
    }
  }
}

// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Hindi (`hi`).
class AppLocalizationsHi extends AppLocalizations {
  AppLocalizationsHi([String locale = 'hi']) : super(locale);

  @override
  String get app_title => 'ज़िप्पायरा स्टाफ ऐप';

  @override
  String get auth_login_title => 'स्टाफ साइन इन';

  @override
  String get auth_phone_subtitle =>
      'आगे बढ़ने के लिए अपना पंजीकृत मोबाइल नंबर दर्ज करें';

  @override
  String get auth_send_otp => 'ओटीपी भेजें';

  @override
  String get auth_verify_otp => 'ओटीपी सत्यापित करें';

  @override
  String get auth_otp_sent_to => 'भेजा गया 6-अंकों का ओटीपी दर्ज करें';

  @override
  String get auth_resend_otp => 'ओटीपी पुनः भेजें';

  @override
  String get auth_staff_not_registered =>
      'यह नंबर स्टाफ के रूप में पंजीकृत नहीं है। अपने स्टोर मैनेजर से जोड़ने के लिए कहें।';

  @override
  String get auth_use_pin => 'पिन का उपयोग करें';

  @override
  String get auth_use_otp_instead => 'ओटीपी का उपयोग करें';

  @override
  String get auth_pin_setup_title => 'क्विक पिन सेट करें';

  @override
  String get auth_pin_setup_later => 'बाद में सेट करें';

  @override
  String get auth_pin_locked_retry =>
      'विफल प्रयासों के कारण पिन लॉक हो गया है। ओटीपी से लॉगिन करें।';

  @override
  String get auth_pin_not_set_fallback =>
      'कोई पिन सेट नहीं है। ओटीपी लॉगिन का उपयोग किया जा रहा है।';

  @override
  String get shift_tab_title => 'शिफ्ट और होम';

  @override
  String get shift_start_shift => 'शिफ्ट शुरू करें';

  @override
  String get shift_end_shift => 'शिफ्ट समाप्त करें';

  @override
  String get shift_duration => 'सक्रिय शिफ्ट अवधि';

  @override
  String get shift_quick_actions => 'त्वरित कार्य';

  @override
  String get shift_already_active => 'शिफ्ट पहले से ही सक्रिय है।';

  @override
  String get shift_ended => 'शिफ्ट समाप्त हो गई।';

  @override
  String get inventory_tab_title => 'इन्वेंट्री';

  @override
  String get inventory_low_stock_title => 'कम स्टॉक अलर्ट';

  @override
  String get inventory_stock_count_title => 'स्टॉक गणना';

  @override
  String get inventory_grn_title => 'जीआरएन प्राप्त करें';

  @override
  String get inventory_submit_count => 'गणना सबमिट करें';

  @override
  String get inventory_count_saved_offline =>
      'ऑफलाइन सहेजा गया — कनेक्ट होने पर सिंक होगा';

  @override
  String get inventory_variance_found => 'अंतर पाया गया';

  @override
  String get grn_title => 'गुड्स रिसीव्ड नोट्स';

  @override
  String get grn_create_adhoc => 'तदर्थ जीआरएन बनाएं';

  @override
  String get grn_qc_review => 'क्यूसी समीक्षा';

  @override
  String get grn_qc_incomplete => 'क्यूसी अपूर्ण — निर्णयों की आवश्यकता है';

  @override
  String get grn_already_completed => 'जीआरएन पहले ही पूर्ण हो चुका है';

  @override
  String get pos_assist_tab_title => 'पीओएस सहायता';

  @override
  String get pos_cash_payment_title => 'नकद भुगतान';

  @override
  String get pos_cash_collected => 'प्राप्त नकद (₹)';

  @override
  String get pos_change_due => 'वापसी राशि';

  @override
  String get pos_insufficient_cash => 'अपर्याप्त नकद प्राप्त';

  @override
  String get pos_payment_confirmed => 'भुगतान की पुष्टि हो गई';

  @override
  String get devices_tab_title => 'उपकरण';

  @override
  String get devices_title => 'स्टोर हार्डवेयर उपकरण';

  @override
  String get devices_status_active => 'सक्रिय';

  @override
  String get devices_status_offline => 'ऑफ़लाइन';

  @override
  String get devices_status_provisioning => 'प्रावधान में';

  @override
  String get devices_last_seen => 'अंतिम बार देखा गया';

  @override
  String get devices_alerts_title => 'अनसुलझे हार्डवेयर अलर्ट';

  @override
  String get devices_resolve => 'समाधान करें';

  @override
  String get devices_exit_attempts_title => 'गेट निकास और आरएफआईडी ऑडिट';

  @override
  String get devices_staff_override => 'स्टाफ ओवरराइड';

  @override
  String get devices_override_reason => 'ओवरराइड का कारण';

  @override
  String get profile_tab_title => 'प्रोफ़ाइल';

  @override
  String get common_coming_soon => 'शीघ्र उपलब्ध होगा';

  @override
  String get common_unauthorized_role =>
      'आपकी भूमिका के पास इस अनुभाग तक पहुंचने की अनुमति नहीं है';

  @override
  String get pairing_title => 'स्टाफ डिवाइस पेयर करें';

  @override
  String get pairing_enter_code =>
      'एडमिनिस्ट्रेटर द्वारा एडमिन प्लेटफॉर्म में जनरेट किया गया 8-अंकों का पेयरिंग कोड दर्ज करें।';

  @override
  String get pairing_invalid_code => 'अमान्य या समाप्त हो चुका पेयरिंग कोड';

  @override
  String get pairing_expired_code =>
      'पेयरिंग कोड समाप्त हो गया है। कृपया एडमिन से नया कोड मांगें।';

  @override
  String get price_check_title => 'कीमत जांचें';

  @override
  String get price_check_not_found =>
      'उत्पाद नहीं मिला। पुन: स्कैन करें या अपने मैनेजर से संपर्क करें।';

  @override
  String get customer_assist_title => 'ग्राहक सहायता';

  @override
  String get customer_assist_enter_last4 =>
      'फ़ोन नंबर के अंतिम 4 अंक दर्ज करें';

  @override
  String get customer_assist_no_match =>
      'इस अंतिम अंक वाले फ़ोन के लिए कोई सक्रिय ग्राहक नहीं मिला:';

  @override
  String get customer_assist_multiple_matches => 'अनेक परिणाम — ग्राहक चुनें';

  @override
  String get customer_assist_scope_note =>
      'यह खोज आपके स्टोर पर पिछले 2 घंटों के सक्रिय ग्राहक सत्रों और ऑर्डर तक सीमित है।';
}

import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_hi.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('hi')
  ];

  /// No description provided for @app_title.
  ///
  /// In en, this message translates to:
  /// **'Zippyra Staff App'**
  String get app_title;

  /// No description provided for @auth_login_title.
  ///
  /// In en, this message translates to:
  /// **'Staff Sign In'**
  String get auth_login_title;

  /// No description provided for @auth_phone_subtitle.
  ///
  /// In en, this message translates to:
  /// **'Enter your registered mobile number to continue'**
  String get auth_phone_subtitle;

  /// No description provided for @auth_send_otp.
  ///
  /// In en, this message translates to:
  /// **'Send OTP'**
  String get auth_send_otp;

  /// No description provided for @auth_verify_otp.
  ///
  /// In en, this message translates to:
  /// **'Verify OTP'**
  String get auth_verify_otp;

  /// No description provided for @auth_otp_sent_to.
  ///
  /// In en, this message translates to:
  /// **'Enter the 6-digit OTP sent to'**
  String get auth_otp_sent_to;

  /// No description provided for @auth_resend_otp.
  ///
  /// In en, this message translates to:
  /// **'Resend OTP'**
  String get auth_resend_otp;

  /// No description provided for @auth_staff_not_registered.
  ///
  /// In en, this message translates to:
  /// **'This number isn\'t registered as staff. Ask your store manager to add you.'**
  String get auth_staff_not_registered;

  /// No description provided for @auth_use_pin.
  ///
  /// In en, this message translates to:
  /// **'Use PIN instead'**
  String get auth_use_pin;

  /// No description provided for @auth_use_otp_instead.
  ///
  /// In en, this message translates to:
  /// **'Use OTP instead'**
  String get auth_use_otp_instead;

  /// No description provided for @auth_pin_setup_title.
  ///
  /// In en, this message translates to:
  /// **'Set up Quick PIN'**
  String get auth_pin_setup_title;

  /// No description provided for @auth_pin_setup_later.
  ///
  /// In en, this message translates to:
  /// **'Set up later'**
  String get auth_pin_setup_later;

  /// No description provided for @auth_pin_locked_retry.
  ///
  /// In en, this message translates to:
  /// **'PIN locked due to failed attempts. Try OTP login instead.'**
  String get auth_pin_locked_retry;

  /// No description provided for @auth_pin_not_set_fallback.
  ///
  /// In en, this message translates to:
  /// **'No PIN configured yet. Falling back to OTP login.'**
  String get auth_pin_not_set_fallback;

  /// No description provided for @shift_tab_title.
  ///
  /// In en, this message translates to:
  /// **'Shift & Home'**
  String get shift_tab_title;

  /// No description provided for @shift_start_shift.
  ///
  /// In en, this message translates to:
  /// **'Start Shift'**
  String get shift_start_shift;

  /// No description provided for @shift_end_shift.
  ///
  /// In en, this message translates to:
  /// **'End Shift'**
  String get shift_end_shift;

  /// No description provided for @shift_duration.
  ///
  /// In en, this message translates to:
  /// **'Active Shift Duration'**
  String get shift_duration;

  /// No description provided for @shift_quick_actions.
  ///
  /// In en, this message translates to:
  /// **'Quick Actions'**
  String get shift_quick_actions;

  /// No description provided for @shift_already_active.
  ///
  /// In en, this message translates to:
  /// **'Shift is already active.'**
  String get shift_already_active;

  /// No description provided for @shift_ended.
  ///
  /// In en, this message translates to:
  /// **'Shift ended.'**
  String get shift_ended;

  /// No description provided for @inventory_tab_title.
  ///
  /// In en, this message translates to:
  /// **'Inventory'**
  String get inventory_tab_title;

  /// No description provided for @inventory_low_stock_title.
  ///
  /// In en, this message translates to:
  /// **'Low Stock Alerts'**
  String get inventory_low_stock_title;

  /// No description provided for @inventory_stock_count_title.
  ///
  /// In en, this message translates to:
  /// **'Stock Count'**
  String get inventory_stock_count_title;

  /// No description provided for @inventory_grn_title.
  ///
  /// In en, this message translates to:
  /// **'GRN Receive'**
  String get inventory_grn_title;

  /// No description provided for @inventory_submit_count.
  ///
  /// In en, this message translates to:
  /// **'Submit Count'**
  String get inventory_submit_count;

  /// No description provided for @inventory_count_saved_offline.
  ///
  /// In en, this message translates to:
  /// **'Saved offline — will sync when connected'**
  String get inventory_count_saved_offline;

  /// No description provided for @inventory_variance_found.
  ///
  /// In en, this message translates to:
  /// **'Discrepancies Found'**
  String get inventory_variance_found;

  /// No description provided for @grn_title.
  ///
  /// In en, this message translates to:
  /// **'Goods Received Notes'**
  String get grn_title;

  /// No description provided for @grn_create_adhoc.
  ///
  /// In en, this message translates to:
  /// **'Create Ad-hoc GRN'**
  String get grn_create_adhoc;

  /// No description provided for @grn_qc_review.
  ///
  /// In en, this message translates to:
  /// **'QC Review'**
  String get grn_qc_review;

  /// No description provided for @grn_qc_incomplete.
  ///
  /// In en, this message translates to:
  /// **'QC Incomplete — decisions needed'**
  String get grn_qc_incomplete;

  /// No description provided for @grn_already_completed.
  ///
  /// In en, this message translates to:
  /// **'GRN Already Completed'**
  String get grn_already_completed;

  /// No description provided for @pos_assist_tab_title.
  ///
  /// In en, this message translates to:
  /// **'POS Assist'**
  String get pos_assist_tab_title;

  /// No description provided for @pos_cash_payment_title.
  ///
  /// In en, this message translates to:
  /// **'Cash Payment'**
  String get pos_cash_payment_title;

  /// No description provided for @pos_cash_collected.
  ///
  /// In en, this message translates to:
  /// **'Cash Collected (₹)'**
  String get pos_cash_collected;

  /// No description provided for @pos_change_due.
  ///
  /// In en, this message translates to:
  /// **'Change Due'**
  String get pos_change_due;

  /// No description provided for @pos_insufficient_cash.
  ///
  /// In en, this message translates to:
  /// **'Insufficient cash collected'**
  String get pos_insufficient_cash;

  /// No description provided for @pos_payment_confirmed.
  ///
  /// In en, this message translates to:
  /// **'Payment Confirmed'**
  String get pos_payment_confirmed;

  /// No description provided for @devices_tab_title.
  ///
  /// In en, this message translates to:
  /// **'Devices'**
  String get devices_tab_title;

  /// No description provided for @devices_title.
  ///
  /// In en, this message translates to:
  /// **'Store Hardware Devices'**
  String get devices_title;

  /// No description provided for @devices_status_active.
  ///
  /// In en, this message translates to:
  /// **'ACTIVE'**
  String get devices_status_active;

  /// No description provided for @devices_status_offline.
  ///
  /// In en, this message translates to:
  /// **'OFFLINE'**
  String get devices_status_offline;

  /// No description provided for @devices_status_provisioning.
  ///
  /// In en, this message translates to:
  /// **'PROVISIONING'**
  String get devices_status_provisioning;

  /// No description provided for @devices_last_seen.
  ///
  /// In en, this message translates to:
  /// **'Last seen'**
  String get devices_last_seen;

  /// No description provided for @devices_alerts_title.
  ///
  /// In en, this message translates to:
  /// **'Unresolved Hardware Alerts'**
  String get devices_alerts_title;

  /// No description provided for @devices_resolve.
  ///
  /// In en, this message translates to:
  /// **'Resolve'**
  String get devices_resolve;

  /// No description provided for @devices_exit_attempts_title.
  ///
  /// In en, this message translates to:
  /// **'Gate Exit & RFID Audit Monitor'**
  String get devices_exit_attempts_title;

  /// No description provided for @devices_staff_override.
  ///
  /// In en, this message translates to:
  /// **'Staff Override'**
  String get devices_staff_override;

  /// No description provided for @devices_override_reason.
  ///
  /// In en, this message translates to:
  /// **'Override Reason'**
  String get devices_override_reason;

  /// No description provided for @profile_tab_title.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile_tab_title;

  /// No description provided for @common_coming_soon.
  ///
  /// In en, this message translates to:
  /// **'Coming Soon'**
  String get common_coming_soon;

  /// No description provided for @common_unauthorized_role.
  ///
  /// In en, this message translates to:
  /// **'Your role does not have permission to access this section'**
  String get common_unauthorized_role;

  /// No description provided for @pairing_title.
  ///
  /// In en, this message translates to:
  /// **'Pair Staff Device'**
  String get pairing_title;

  /// No description provided for @pairing_enter_code.
  ///
  /// In en, this message translates to:
  /// **'Enter the 8-character pairing code generated by your Administrator in the Zippyra Admin Platform.'**
  String get pairing_enter_code;

  /// No description provided for @pairing_invalid_code.
  ///
  /// In en, this message translates to:
  /// **'Invalid or expired pairing code'**
  String get pairing_invalid_code;

  /// No description provided for @pairing_expired_code.
  ///
  /// In en, this message translates to:
  /// **'Pairing code expired. Please request a new code from Admin.'**
  String get pairing_expired_code;

  /// No description provided for @price_check_title.
  ///
  /// In en, this message translates to:
  /// **'Price Check'**
  String get price_check_title;

  /// No description provided for @price_check_not_found.
  ///
  /// In en, this message translates to:
  /// **'Product not found. Try scanning again or check with your manager.'**
  String get price_check_not_found;

  /// No description provided for @customer_assist_title.
  ///
  /// In en, this message translates to:
  /// **'Customer Assist'**
  String get customer_assist_title;

  /// No description provided for @customer_assist_enter_last4.
  ///
  /// In en, this message translates to:
  /// **'Enter Last 4 Digits of Phone Number'**
  String get customer_assist_enter_last4;

  /// No description provided for @customer_assist_no_match.
  ///
  /// In en, this message translates to:
  /// **'No active customer found for phone ending in'**
  String get customer_assist_no_match;

  /// No description provided for @customer_assist_multiple_matches.
  ///
  /// In en, this message translates to:
  /// **'Multiple Matches - Select Customer'**
  String get customer_assist_multiple_matches;

  /// No description provided for @customer_assist_scope_note.
  ///
  /// In en, this message translates to:
  /// **'This lookup is scoped to active customer sessions and orders at your store within the last 2 hours for support purposes.'**
  String get customer_assist_scope_note;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'hi'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'hi':
      return AppLocalizationsHi();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}

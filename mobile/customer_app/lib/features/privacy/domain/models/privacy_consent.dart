class PrivacyConsent {
  final String consentType;
  final bool granted;
  final String consentVersion;
  final bool needsReconfirmation;

  const PrivacyConsent({
    required this.consentType,
    required this.granted,
    required this.consentVersion,
    this.needsReconfirmation = false,
  });

  PrivacyConsent copyWith({
    String? consentType,
    bool? granted,
    String? consentVersion,
    bool? needsReconfirmation,
  }) {
    return PrivacyConsent(
      consentType: consentType ?? this.consentType,
      granted: granted ?? this.granted,
      consentVersion: consentVersion ?? this.consentVersion,
      needsReconfirmation: needsReconfirmation ?? this.needsReconfirmation,
    );
  }

  factory PrivacyConsent.fromJson(Map<String, dynamic> json) {
    return PrivacyConsent(
      consentType: json['consent_type'] as String? ?? '',
      granted: json['granted'] as bool? ?? false,
      consentVersion: json['consent_version'] as String? ?? 'v1.0',
      needsReconfirmation: json['needs_reconfirmation'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'consent_type': consentType,
        'granted': granted,
        'consent_version': consentVersion,
        'needs_reconfirmation': needsReconfirmation,
      };
}

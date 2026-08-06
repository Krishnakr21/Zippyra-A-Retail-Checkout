class User {
  final String id;
  final String? phone;
  final String? email;
  final String? googleSub;
  final DateTime? emailVerifiedAt;
  final DateTime? phoneVerifiedAt;
  final String? authProviderLast;

  const User({
    required this.id,
    this.phone,
    this.email,
    this.googleSub,
    this.emailVerifiedAt,
    this.phoneVerifiedAt,
    this.authProviderLast,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String? ?? '',
      phone: json['phone'] as String?,
      email: json['email'] as String?,
      googleSub: json['google_sub'] as String?,
      emailVerifiedAt: json['email_verified_at'] != null ? DateTime.tryParse(json['email_verified_at']) : null,
      phoneVerifiedAt: json['phone_verified_at'] != null ? DateTime.tryParse(json['phone_verified_at']) : null,
      authProviderLast: json['auth_provider_last'] as String?,
    );
  }
}

class AuthSession {
  final String accessToken;
  final String refreshToken;
  final User user;
  final bool isNewUser;

  const AuthSession({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
    required this.isNewUser,
  });

  factory AuthSession.fromJson(Map<String, dynamic> json) {
    return AuthSession(
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
      user: User.fromJson(json['user'] as Map<String, dynamic>? ?? {}),
      isNewUser: json['is_new_user'] as bool? ?? false,
    );
  }
}

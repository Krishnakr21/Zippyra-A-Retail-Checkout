class StoreSession {
  final String storeId;
  final String storeName;
  final String sessionToken;
  final int catalogVersion;
  final String expiresAt;
  final bool rfidEnabled;

  const StoreSession({
    required this.storeId,
    required this.storeName,
    required this.sessionToken,
    required this.catalogVersion,
    required this.expiresAt,
    this.rfidEnabled = false,
  });
}

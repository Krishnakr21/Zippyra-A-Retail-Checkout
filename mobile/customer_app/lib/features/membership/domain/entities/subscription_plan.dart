class SubscriptionPlan {
  final String id;
  final String chainId;
  final String name;
  final int pricePaise;
  final String billingInterval; // MONTHLY | ANNUAL
  final double loyaltyMultiplierBonus;
  final bool freeDelivery;
  final bool isActive;

  const SubscriptionPlan({
    required this.id,
    required this.chainId,
    required this.name,
    required this.pricePaise,
    required this.billingInterval,
    required this.loyaltyMultiplierBonus,
    required this.freeDelivery,
    required this.isActive,
  });

  String get formattedPrice {
    final rupee = (pricePaise / 100).toStringAsFixed(0);
    return '₹$rupee / ${billingInterval == 'ANNUAL' ? 'year' : 'month'}';
  }
}

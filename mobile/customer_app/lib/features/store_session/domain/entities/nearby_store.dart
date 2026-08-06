class NearbyStore {
  final String id;
  final String name;
  final String address;
  final double distanceKm;
  final bool isOpen;
  final int capacityPct;

  const NearbyStore({
    required this.id,
    required this.name,
    required this.address,
    required this.distanceKm,
    required this.isOpen,
    required this.capacityPct,
  });
}

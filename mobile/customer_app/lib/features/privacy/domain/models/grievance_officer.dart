class GrievanceOfficer {
  final String name;
  final String title;
  final String email;
  final String address;
  final String acknowledgmentSla;

  const GrievanceOfficer({
    required this.name,
    required this.title,
    required this.email,
    required this.address,
    required this.acknowledgmentSla,
  });

  factory GrievanceOfficer.fromJson(Map<String, dynamic> json) {
    return GrievanceOfficer(
      name: json['name'] as String? ?? 'Nisha Sharma',
      title: json['title'] as String? ?? 'Data Protection & Grievance Officer',
      email: json['email'] as String? ?? 'grievance@zippyra.com',
      address: json['address'] as String? ?? 'Zippyra India Tech Pvt Ltd, Bengaluru 560102',
      acknowledgmentSla: json['acknowledgment_sla'] as String? ?? '72 hours',
    );
  }
}

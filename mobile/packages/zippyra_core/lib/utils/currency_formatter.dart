class CurrencyFormatter {
  static String formatPaise(int paise) {
    final rupees = paise / 100.0;
    return '₹${rupees.toStringAsFixed(2)}';
  }
}

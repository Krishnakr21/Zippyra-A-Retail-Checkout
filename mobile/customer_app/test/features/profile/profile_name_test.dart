import 'package:flutter_test/flutter_test.dart';

void main() {
  test('Profile name validation rejects empty or whitespace name', () {
    String validateName(String input) {
      final trimmed = input.trim();
      if (trimmed.isEmpty) return 'Display name cannot be empty';
      if (trimmed.length > 100) return 'Name exceeds 100 characters';
      return 'VALID';
    }

    expect(validateName(''), equals('Display name cannot be empty'));
    expect(validateName('   '), equals('Display name cannot be empty'));
    expect(validateName('Anita Sharma'), equals('VALID'));
  });
}

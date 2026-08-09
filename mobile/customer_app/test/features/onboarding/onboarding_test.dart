import 'package:flutter_test/flutter_test.dart';

class FakeStorage {
  final Map<String, String> _data = {};

  Future<String?> read(String key) async => _data[key];
  Future<void> write(String key, String value) async => _data[key] = value;
}

void main() {
  late FakeStorage storage;

  setUp(() {
    storage = FakeStorage();
  });

  test('Onboarding flag is empty initially for first launch', () async {
    final flag = await storage.read('has_seen_onboarding');
    expect(flag, isNull);
  });

  test('Completing onboarding persists has_seen_onboarding flag as true', () async {
    await storage.write('has_seen_onboarding', 'true');
    final flag = await storage.read('has_seen_onboarding');
    expect(flag, equals('true'));
  });
}

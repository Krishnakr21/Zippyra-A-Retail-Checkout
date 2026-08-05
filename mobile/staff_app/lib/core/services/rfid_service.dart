import 'dart:async';

abstract class RfidService {
  Future<void> connect();
  Stream<String> get rfidStatusStream;
  void dispose();
}

class RfidServiceImpl implements RfidService {
  final _statusController = StreamController<String>.broadcast();
  bool _isConnected = false;

  @override
  Stream<String> get rfidStatusStream => _statusController.stream;

  @override
  Future<void> connect() async {
    _isConnected = true;
    _statusController.add('CONNECTED');
  }

  @override
  void dispose() {
    _statusController.close();
  }
}

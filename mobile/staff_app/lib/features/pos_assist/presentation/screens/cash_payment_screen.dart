import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';

class CashPaymentScreen extends StatefulWidget {
  const CashPaymentScreen({super.key});

  @override
  State<CashPaymentScreen> createState() => _CashPaymentScreenState();
}

class _CashPaymentScreenState extends State<CashPaymentScreen> {
  final _sessionController = TextEditingController();
  final _cashController = TextEditingController();
  int _totalPaise = 50000; // Sample checkout total ₹500.00
  bool _isProcessing = false;
  Map<String, dynamic>? _successResult;

  @override
  void dispose() {
    _sessionController.dispose();
    _cashController.dispose();
    super.dispose();
  }

  void _onNumPadTap(String val) {
    if (val == 'C') {
      _cashController.clear();
    } else if (val == '⌫') {
      final text = _cashController.text;
      if (text.isNotEmpty) {
        _cashController.text = text.substring(0, text.length - 1);
      }
    } else {
      _cashController.text += val;
    }
    setState(() {});
  }

  Future<void> _processCashPayment() async {
    final sessionId = _sessionController.text.trim();
    final cashRupees = double.tryParse(_cashController.text.trim()) ?? 0.0;
    final cashPaise = (cashRupees * 100).toInt();

    if (sessionId.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter Checkout Session ID')),
      );
      return;
    }

    if (cashPaise < _totalPaise) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Insufficient cash collected'), backgroundColor: Colors.red),
      );
      return;
    }

    setState(() => _isProcessing = true);
    await Future.delayed(const Duration(milliseconds: 600)); // Simulate API call to payment-service

    final changeDuePaise = cashPaise - _totalPaise;

    setState(() {
      _isProcessing = false;
      _successResult = {
        'session_id': sessionId,
        'cash_collected_paise': cashPaise,
        'change_due_paise': changeDuePaise,
      };
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_successResult != null) {
      final changeDueRupees = (_successResult!['change_due_paise'] as int) / 100.0;
      return Scaffold(
        backgroundColor: Colors.green[800],
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Icon(Icons.check_circle_outline, size: 96, color: Colors.white),
                const SizedBox(height: 24),
                const Text(
                  'Payment Confirmed',
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: Colors.white),
                ),
                const SizedBox(height: 16),
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Column(
                    children: [
                      const Text('CHANGE DUE TO CUSTOMER', style: TextStyle(color: Colors.grey, fontSize: 12, fontWeight: FontWeight.bold)),
                      const SizedBox(height: 8),
                      Text(
                        '₹${changeDueRupees.toStringAsFixed(2)}',
                        style: const TextStyle(fontSize: 40, fontWeight: FontWeight.bold, color: Colors.green),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 32),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.white,
                    foregroundColor: Colors.green[900],
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                  onPressed: () {
                    setState(() {
                      _successResult = null;
                      _sessionController.clear();
                      _cashController.clear();
                    });
                  },
                  child: const Text('Next Transaction', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                ),
              ],
            ),
          ),
        ),
      );
    }

    final cashRupees = double.tryParse(_cashController.text.trim()) ?? 0.0;
    final cashPaise = (cashRupees * 100).toInt();
    final changeDuePaise = (cashPaise - _totalPaise).clamp(0, 99999999);
    final changeDueRupees = changeDuePaise / 100.0;

    return Scaffold(
      appBar: AppBar(title: const Text('Cash Payment Assist')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ZInput(
              label: 'Session ID',
              hint: 'Checkout Session ID',
              controller: _sessionController,
            ),
            const SizedBox(height: 16),

            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Total Amount Due:', style: TextStyle(fontSize: 16, color: Colors.grey)),
                    Text(
                      '₹${(_totalPaise / 100.0).toStringAsFixed(2)}',
                      style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: ZippyraColors.primary),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 12),

            Card(
              color: Colors.green[50],
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    Text('Cash Collected: ₹${_cashController.text.isEmpty ? '0' : _cashController.text}', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    const Divider(height: 16),
                    Text(
                      'Change Due: ₹${changeDueRupees.toStringAsFixed(2)}',
                      style: TextStyle(
                        fontSize: 22,
                        fontWeight: FontWeight.bold,
                        color: cashPaise >= _totalPaise ? Colors.green[800] : Colors.grey,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // On-Screen Numeric Keypad
            GridView.count(
              crossAxisCount: 3,
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              childAspectRatio: 2.0,
              mainAxisSpacing: 8,
              crossAxisSpacing: 8,
              children: [
                '1', '2', '3',
                '4', '5', '6',
                '7', '8', '9',
                'C', '0', '⌫'
              ].map((key) {
                return OutlinedButton(
                  onPressed: () => _onNumPadTap(key),
                  child: Text(key, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
                );
              }).toList(),
            ),

            const SizedBox(height: 20),
            ZButton(
              label: 'Confirm Cash Payment',
              isLoading: _isProcessing,
              onPressed: _processCashPayment,
            ),
          ],
        ),
      ),
    );
  }
}

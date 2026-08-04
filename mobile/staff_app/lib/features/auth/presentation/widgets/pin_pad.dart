import 'package:flutter/material.dart';

class PinPadWidget extends StatefulWidget {
  final int maxLength;
  final bool isLocked;
  final String? lockedText;
  final ValueChanged<String> onCompleted;

  const PinPadWidget({
    super.key,
    this.maxLength = 4,
    this.isLocked = false,
    this.lockedText,
    required this.onCompleted,
  });

  @override
  State<PinPadWidget> createState() => _PinPadWidgetState();
}

class _PinPadWidgetState extends State<PinPadWidget> {
  String _pin = '';

  void _onKeyPress(String digit) {
    if (widget.isLocked) return;
    if (_pin.length < widget.maxLength) {
      setState(() {
        _pin += digit;
      });
      if (_pin.length == widget.maxLength) {
        widget.onCompleted(_pin);
      }
    }
  }

  void _onBackspace() {
    if (widget.isLocked) return;
    if (_pin.isNotEmpty) {
      setState(() {
        _pin = _pin.substring(0, _pin.length - 1);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // Dots Indicator
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            widget.maxLength,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 8),
              width: 16,
              height: 16,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: index < _pin.length
                    ? Theme.of(context).primaryColor
                    : Colors.grey.shade300,
              ),
            ),
          ),
        ),

        if (widget.isLocked && widget.lockedText != null) ...[
          const SizedBox(height: 16),
          Text(
            widget.lockedText!,
            style: const TextStyle(color: Colors.red, fontWeight: FontWeight.bold, fontSize: 13),
            textAlign: TextAlign.center,
          ),
        ],

        const SizedBox(height: 24),

        // Keypad Grid
        Opacity(
          opacity: widget.isLocked ? 0.4 : 1.0,
          child: Column(
            children: [
              for (var row in [
                ['1', '2', '3'],
                ['4', '5', '6'],
                ['7', '8', '9'],
              ])
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: row.map((digit) => _buildKey(digit)).toList(),
                ),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  const SizedBox(width: 64, height: 64),
                  _buildKey('0'),
                  InkWell(
                    onTap: widget.isLocked ? null : _onBackspace,
                    borderRadius: BorderRadius.circular(32),
                    child: Container(
                      width: 64,
                      height: 64,
                      alignment: Alignment.center,
                      child: const Icon(Icons.backspace_outlined),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildKey(String digit) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: InkWell(
        onTap: widget.isLocked ? null : () => _onKeyPress(digit),
        borderRadius: BorderRadius.circular(32),
        child: Container(
          width: 64,
          height: 64,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: Colors.grey.shade100,
          ),
          child: Text(
            digit,
            style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
          ),
        ),
      ),
    );
  }
}

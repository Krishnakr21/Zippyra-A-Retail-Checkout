import 'package:flutter/material.dart';
import 'package:zippyra_core/zippyra_core.dart';

class CouponInputField extends StatefulWidget {
  final String? appliedCoupon;
  final ValueChanged<String> onApply;
  final VoidCallback onRemove;
  final String? errorMessage;

  const CouponInputField({
    super.key,
    this.appliedCoupon,
    required this.onApply,
    required this.onRemove,
    this.errorMessage,
  });

  @override
  State<CouponInputField> createState() => _CouponInputFieldState();
}

class _CouponInputFieldState extends State<CouponInputField> {
  late TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.appliedCoupon ?? '');
  }

  @override
  void didUpdateWidget(covariant CouponInputField oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.appliedCoupon != oldWidget.appliedCoupon && widget.appliedCoupon != null) {
      _controller.text = widget.appliedCoupon!;
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final hasCoupon = widget.appliedCoupon != null && widget.appliedCoupon!.isNotEmpty;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _controller,
                  enabled: !hasCoupon,
                  textCapitalization: TextCapitalization.characters,
                  decoration: InputDecoration(
                    hintText: 'Enter Coupon Code (e.g. SAVE50)',
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                    suffixIcon: hasCoupon
                        ? const Icon(Icons.check_circle, color: Colors.green)
                        : null,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: hasCoupon ? Colors.red : ZippyraColors.primary,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                ),
                onPressed: () {
                  if (hasCoupon) {
                    _controller.clear();
                    widget.onRemove();
                  } else {
                    final code = _controller.text.trim();
                    if (code.isNotEmpty) {
                      widget.onApply(code);
                    }
                  }
                },
                child: Text(
                  hasCoupon ? 'Remove' : 'Apply',
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
          if (widget.errorMessage != null && widget.errorMessage!.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4, left: 4),
              child: Text(
                widget.errorMessage!,
                style: const TextStyle(color: Colors.red, fontSize: 12),
              ),
            ),
        ],
      ),
    );
  }
}

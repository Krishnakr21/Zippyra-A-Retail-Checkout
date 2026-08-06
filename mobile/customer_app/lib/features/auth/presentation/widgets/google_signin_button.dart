import 'package:flutter/material.dart';

class GoogleSignInButton extends StatelessWidget {
  final VoidCallback onPressed;
  final bool isLoading;

  const GoogleSignInButton({
    super.key,
    required this.onPressed,
    this.isLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: isLoading ? null : onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
        shadowColor: Colors.black26,
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8.0),
          side: const BorderSide(color: Color(0xFFDADCE0), width: 1.0),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (isLoading)
            const SizedBox(
              height: 20,
              width: 20,
              child: CircularProgressIndicator(strokeWidth: 2, valueColor: AlwaysStoppedAnimation<Color>(Colors.blue)),
            )
          else ...[
            // Google 'G' Icon
            Container(
              height: 20,
              width: 20,
              decoration: const BoxDecoration(
                shape: BoxShape.circle,
              ),
              child: CustomPaint(
                painter: _GoogleGLogoPainter(),
              ),
            ),
            const SizedBox(width: 12),
            const Text(
              'Continue with Google',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
                color: Color(0xFF3C4043),
                letterSpacing: 0.25,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _GoogleGLogoPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final double cx = size.width / 2;
    final double cy = size.height / 2;
    final double radius = size.width / 2;

    final Paint redPaint = Paint()..color = const Color(0xFFEA4335);
    final Paint bluePaint = Paint()..color = const Color(0xFF4285F4);
    final Paint greenPaint = Paint()..color = const Color(0xFF34A853);
    final Paint yellowPaint = Paint()..color = const Color(0xFFFBBC05);

    // Draw Google G shapes
    final Rect rect = Rect.fromCircle(center: Offset(cx, cy), radius: radius);

    canvas.drawArc(rect, -0.4, 1.9, true, redPaint);
    canvas.drawArc(rect, 1.5, 1.1, true, yellowPaint);
    canvas.drawArc(rect, 2.6, 1.2, true, greenPaint);
    canvas.drawArc(rect, 3.8, 1.6, true, bluePaint);

    final Paint whiteInner = Paint()..color = Colors.white;
    canvas.drawCircle(Offset(cx, cy), radius * 0.55, whiteInner);

    final Path blueBar = Path()
      ..addRect(Rect.fromLTRB(cx - 2, cy - radius * 0.25, cx + radius, cy + radius * 0.25));
    canvas.drawPath(blueBar, bluePaint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

import 'dart:async';
import 'package:razorpay_flutter/razorpay_flutter.dart';

class RazorpayService {
  late Razorpay _razorpay;
  final StreamController<PaymentSuccessResponse> _successController = StreamController<PaymentSuccessResponse>.broadcast();
  final StreamController<PaymentFailureResponse> _errorController = StreamController<PaymentFailureResponse>.broadcast();
  final StreamController<ExternalWalletResponse> _walletController = StreamController<ExternalWalletResponse>.broadcast();

  Stream<PaymentSuccessResponse> get onSuccess => _successController.stream;
  Stream<PaymentFailureResponse> get onError => _errorController.stream;
  Stream<ExternalWalletResponse> get onExternalWallet => _walletController.stream;

  RazorpayService() {
    _razorpay = Razorpay();
    _razorpay.on(Razorpay.EVENT_PAYMENT_SUCCESS, _handlePaymentSuccess);
    _razorpay.on(Razorpay.EVENT_PAYMENT_ERROR, _handlePaymentError);
    _razorpay.on(Razorpay.EVENT_EXTERNAL_WALLET, _handleExternalWallet);
  }

  void _handlePaymentSuccess(PaymentSuccessResponse response) {
    _successController.add(response);
  }

  void _handlePaymentError(PaymentFailureResponse response) {
    _errorController.add(response);
  }

  void _handleExternalWallet(ExternalWalletResponse response) {
    _walletController.add(response);
  }

  void open(Map<String, dynamic> options) {
    _razorpay.open(options);
  }

  void clear() {
    _razorpay.clear();
  }

  void dispose() {
    _razorpay.clear();
    _successController.close();
    _errorController.close();
    _walletController.close();
  }
}

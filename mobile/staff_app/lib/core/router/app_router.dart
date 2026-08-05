import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../widgets/main_shell_scaffold.dart';
import '../widgets/role_guard.dart';
import '../../features/auth/presentation/screens/login_screen.dart';
import '../../features/auth/presentation/screens/otp_screen.dart';
import '../../features/inventory/presentation/screens/low_stock_screen.dart';
import '../../features/inventory/presentation/screens/stock_count_screen.dart';
import '../../features/inventory/presentation/screens/grn_list_screen.dart';
import '../../features/inventory/presentation/screens/grn_receive_screen.dart';
import '../../features/inventory/presentation/screens/qc_review_screen.dart';
import '../../features/pos_assist/presentation/screens/cash_payment_screen.dart';
import '../../features/pos_assist/presentation/screens/price_check_screen.dart';
import '../../features/devices/presentation/screens/devices_screen.dart';
import '../../features/customer_assist/presentation/screens/customer_assist_screen.dart';
import '../../features/profile/presentation/screens/profile_screen.dart';

final GoRouter staffAppRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/otp',
      builder: (context, state) {
        final phone = state.extra as String? ?? '';
        return OtpScreen(phone: phone);
      },
    ),
    GoRoute(
      path: '/home/shift',
      builder: (context, state) => const MainShellScaffold(initialIndex: 0),
    ),
    GoRoute(
      path: '/home/inventory',
      builder: (context, state) => const MainShellScaffold(initialIndex: 1),
    ),
    GoRoute(
      path: '/home/inventory/low-stock',
      builder: (context, state) => const LowStockScreen(),
    ),
    GoRoute(
      path: '/home/inventory/stock-count',
      builder: (context, state) => const StockCountScreen(),
    ),
    GoRoute(
      path: '/home/inventory/grn',
      builder: (context, state) => const RoleGuard(
        allowedRoles: ['STOCK_ASSOCIATE', 'MANAGER'],
        child: GrnListScreen(),
      ),
    ),
    GoRoute(
      path: '/home/inventory/grn/:id',
      builder: (context, state) {
        final id = state.pathParameters['id'];
        return RoleGuard(
          allowedRoles: ['STOCK_ASSOCIATE', 'MANAGER'],
          child: GrnReceiveScreen(poId: id),
        );
      },
    ),
    GoRoute(
      path: '/home/inventory/qc/:grnId',
      builder: (context, state) {
        final grnId = state.pathParameters['grnId'] ?? '';
        return QcReviewScreen(grnId: grnId);
      },
    ),
    GoRoute(
      path: '/home/devices',
      builder: (context, state) => const MainShellScaffold(initialIndex: 2),
    ),
    GoRoute(
      path: '/home/pos-assist',
      builder: (context, state) => const MainShellScaffold(initialIndex: 3),
    ),
    GoRoute(
      path: '/home/pos-assist/cash-payment',
      builder: (context, state) => const RoleGuard(
        allowedRoles: ['CASHIER', 'MANAGER'],
        child: CashPaymentScreen(),
      ),
    ),
    GoRoute(
      path: '/home/pos-assist/price-check',
      builder: (context, state) => const PriceCheckScreen(),
    ),
    GoRoute(
      path: '/home/customer-assist',
      builder: (context, state) => const CustomerAssistScreen(),
    ),
    GoRoute(
      path: '/home/profile',
      builder: (context, state) => const MainShellScaffold(initialIndex: 4),
    ),
  ],
);

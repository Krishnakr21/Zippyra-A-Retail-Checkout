import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../injection_container.dart';
import '../../features/splash/presentation/splash_screen.dart';
import '../../features/splash/presentation/update_required_screen.dart';
import '../../features/onboarding/presentation/onboarding_screen.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/otp_screen.dart';
import '../../features/auth/presentation/bloc/auth_bloc.dart';
import '../../features/permissions/presentation/permissions_screen.dart';

import '../../features/store_session/presentation/bloc/store_session_bloc.dart';
import '../../features/store_session/presentation/screens/store_list_screen.dart';
import '../../features/store_session/presentation/screens/entrance_scan_screen.dart';
import '../../features/store_session/presentation/screens/store_binding_screen.dart';
import '../../features/store_session/presentation/screens/store_bound_screen.dart';
import '../../features/store_session/presentation/screens/store_closed_screen.dart';
import '../../features/store_session/presentation/screens/store_at_capacity_screen.dart';
import '../../features/store_session/presentation/screens/geofence_mismatch_screen.dart';

import '../../features/catalog/domain/entities/product.dart';
import '../../features/catalog/presentation/bloc/catalog_bloc.dart';
import '../../features/catalog/presentation/screens/search_screen.dart';
import '../../features/catalog/presentation/screens/category_browse_screen.dart';
import '../../features/catalog/presentation/screens/category_products_screen.dart';
import '../../features/catalog/presentation/screens/product_detail_screen.dart';

import '../../features/home/presentation/home_screen.dart';
import '../../features/scan/presentation/scan_screen.dart';
import '../../features/cart/presentation/bloc/cart_bloc.dart';
import '../../features/cart/presentation/screens/cart_screen.dart';
import '../../features/payment/presentation/bloc/payment_bloc.dart';
import '../../features/payment/presentation/screens/checkout_screen.dart';
import '../../features/payment/presentation/screens/payment_failed_screen.dart';
import '../../features/payment/presentation/screens/payment_pending_timeout_screen.dart';
import '../../features/payment/presentation/screens/payment_processing_screen.dart';
import '../../features/payment/presentation/screens/payment_success_screen.dart';
import '../../features/orders/domain/entities/order_detail.dart';
import '../../features/orders/presentation/bloc/order_detail_bloc.dart';
import '../../features/orders/presentation/bloc/order_history_bloc.dart';
import '../../features/orders/presentation/cubit/order_exit_cubit.dart';
import '../../features/orders/presentation/screens/order_detail_screen.dart';
import '../../features/orders/presentation/screens/order_history_screen.dart';
import '../../features/orders/presentation/screens/return_confirmation_screen.dart';
import '../../features/orders/presentation/screens/return_request_screen.dart';
import '../../features/exit/presentation/bloc/exit_bloc.dart';
import '../../features/exit/presentation/screens/exit_expired_screen.dart';
import '../../features/exit/presentation/screens/exit_help_needed_screen.dart';
import '../../features/exit/presentation/screens/exit_qr_screen.dart';
import '../../features/exit/presentation/screens/exit_success_screen.dart';
import '../../features/loyalty/domain/usecases/get_tiers_info_use_case.dart';
import '../../features/loyalty/presentation/cubit/loyalty_history_cubit.dart';
import '../../features/loyalty/presentation/screens/loyalty_history_screen.dart';
import '../../features/loyalty/presentation/screens/loyalty_home_screen.dart';
import '../../features/loyalty/presentation/screens/loyalty_tiers_info_screen.dart';
import '../../features/loyalty/presentation/screens/referral_screen.dart';
import '../../features/loyalty/presentation/cubit/referral_cubit.dart';
import '../../features/exit/presentation/exit_screen.dart';
import '../../features/loyalty/presentation/loyalty_screen.dart';
import '../../features/orders/presentation/orders_screen.dart';
import '../../features/notifications/presentation/notifications_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/profile/presentation/account_edit_screen.dart';
import '../../features/profile/presentation/settings_screen.dart';
import '../../features/profile/presentation/manage_my_data_screen.dart';
import '../../features/profile/presentation/device_sessions_screen.dart';
import '../../features/profile/presentation/bloc/device_sessions_bloc.dart';
import '../../features/membership/presentation/screens/membership_screen.dart';
import '../../features/membership/presentation/cubit/membership_cubit.dart';

final GoRouter appRouter = GoRouter(
  initialLocation: '/splash',
  routes: [
    GoRoute(path: '/splash', builder: (context, state) => const SplashScreen()),
    GoRoute(path: '/update_required', builder: (context, state) => const UpdateRequiredScreen()),
    GoRoute(path: '/onboarding', builder: (context, state) => const OnboardingScreen()),
    GoRoute(
      path: '/auth',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<AuthBloc>(),
        child: const LoginScreen(),
      ),
    ),
    GoRoute(
      path: '/auth/otp',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        final channel = extra['channel'] as String? ?? 'phone';
        final identifier = extra['identifier'] as String? ?? '';
        return BlocProvider(
          create: (_) => sl<AuthBloc>(),
          child: OtpScreen(channel: channel, identifier: identifier),
        );
      },
    ),
    GoRoute(
      path: '/profile-completion',
      builder: (context, state) => const ProfileScreen(),
    ),
    GoRoute(path: '/permissions', builder: (context, state) => const PermissionsScreen()),

    // Store Session Routes (Wrapped with StoreSessionBloc)
    GoRoute(
      path: '/store/list',
      builder: (context, state) => BlocProvider.value(
        value: sl<StoreSessionBloc>(),
        child: const StoreListScreen(),
      ),
    ),
    GoRoute(
      path: '/store/scan',
      builder: (context, state) => BlocProvider.value(
        value: sl<StoreSessionBloc>(),
        child: const EntranceScanScreen(),
      ),
    ),
    GoRoute(
      path: '/store/binding',
      builder: (context, state) => BlocProvider.value(
        value: sl<StoreSessionBloc>(),
        child: const StoreBindingScreen(),
      ),
    ),
    GoRoute(
      path: '/store/bound',
      builder: (context, state) => BlocProvider.value(
        value: sl<StoreSessionBloc>(),
        child: const StoreBoundScreen(),
      ),
    ),
    GoRoute(
      path: '/store/closed',
      builder: (context, state) {
        final msg = state.extra as String?;
        return StoreClosedScreen(message: msg);
      },
    ),
    GoRoute(
      path: '/store/at_capacity',
      builder: (context, state) {
        final msg = state.extra as String?;
        return BlocProvider.value(
          value: sl<StoreSessionBloc>(),
          child: StoreAtCapacityScreen(message: msg),
        );
      },
    ),
    GoRoute(
      path: '/store/geofence_mismatch',
      builder: (context, state) {
        final msg = state.extra as String?;
        return GeofenceMismatchScreen(message: msg);
      },
    ),

    // Catalog Routes (Wrapped with CatalogBloc)
    GoRoute(
      path: '/search',
      builder: (context, state) => const SearchScreen(),
    ),
    GoRoute(
      path: '/catalog/search',
      builder: (context, state) {
        final storeId = state.extra as String? ?? 'store-1';
        return BlocProvider.value(
          value: sl<CatalogBloc>(),
          child: SearchScreen(storeId: storeId),
        );
      },
    ),
    GoRoute(
      path: '/categories',
      builder: (context, state) => const CategoryBrowseScreen(),
    ),
    GoRoute(
      path: '/category/products',
      builder: (context, state) {
        final categoryName = state.uri.queryParameters['name'] ?? state.extra as String? ?? 'All';
        return CategoryProductsScreen(categoryName: categoryName);
      },
    ),
    GoRoute(
      path: '/catalog/categories',
      builder: (context, state) {
        final chainId = state.extra as String? ?? 'chain-hq-001';
        return BlocProvider.value(
          value: sl<CatalogBloc>(),
          child: CategoryBrowseScreen(chainId: chainId),
        );
      },
    ),
    GoRoute(
      path: '/catalog/detail',
      builder: (context, state) {
        final product = state.extra as Product;
        return ProductDetailScreen(product: product);
      },
    ),

    GoRoute(path: '/home', builder: (context, state) => const HomeScreen()),
    GoRoute(
      path: '/scan',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>?;
        return ScanScreen(
          targetProductName: extra?['name'] as String?,
          targetBarcode: extra?['barcode'] as String?,
          targetPrice: extra?['price'] as String?,
          targetImageUrl: extra?['image_url'] as String?,
        );
      },
    ),
    GoRoute(
      path: '/cart',
      builder: (context, state) => BlocProvider.value(
        value: sl<CartBloc>(),
        child: const CartScreen(),
      ),
    ),
    GoRoute(
      path: '/payment/checkout',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        return BlocProvider(
          create: (_) => sl<PaymentBloc>(),
          child: CheckoutScreen(
            checkoutSessionId: extra['checkout_session_id'] as String? ?? 'sess-100',
            totalPaise: (extra['total_paise'] ?? 50000) as int,
          ),
        );
      },
    ),
    GoRoute(
      path: '/payment/processing',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        return BlocProvider(
          create: (_) => sl<PaymentBloc>(),
          child: PaymentProcessingScreen(
            paymentId: extra['payment_id'] as String? ?? '',
          ),
        );
      },
    ),
    GoRoute(
      path: '/payment/success',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        return BlocProvider(
          create: (_) => sl<OrderExitCubit>(),
          child: PaymentSuccessScreen(
            paymentId: extra['payment_id'] as String? ?? '',
            storeId: extra['store_id'] as String? ?? 'store-1',
          ),
        );
      },
    ),
    GoRoute(
      path: '/payment/failed',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        return PaymentFailedScreen(reason: extra['reason'] as String? ?? '');
      },
    ),
    GoRoute(
      path: '/payment/pending-timeout',
      builder: (context, state) => const PaymentPendingTimeoutScreen(),
    ),
    GoRoute(
      path: '/orders',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<OrderHistoryBloc>(),
        child: const OrderHistoryScreen(),
      ),
    ),
    GoRoute(
      path: '/order/:id',
      builder: (context, state) {
        final orderId = state.pathParameters['id'] ?? '';
        return BlocProvider(
          create: (_) => sl<OrderDetailBloc>(),
          child: OrderDetailScreen(orderId: orderId),
        );
      },
    ),
    GoRoute(
      path: '/order/:id/return',
      builder: (context, state) {
        final order = state.extra as OrderDetail;
        return BlocProvider(
          create: (_) => sl<OrderDetailBloc>(),
          child: ReturnRequestScreen(order: order),
        );
      },
    ),
    GoRoute(
      path: '/order/:id/return/confirmation',
      builder: (context, state) => const ReturnConfirmationScreen(),
    ),
    GoRoute(
      path: '/exit/qr',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>? ?? {};
        final orderId = extra['order_id'] as String? ?? '';
        final token = extra['exit_token'] as String? ?? '';
        final expiresAtRaw = extra['expires_at'];
        final DateTime expiresAt = expiresAtRaw is DateTime
            ? expiresAtRaw
            : (expiresAtRaw is String ? DateTime.parse(expiresAtRaw) : DateTime.now().add(const Duration(minutes: 10)));

        return BlocProvider(
          create: (_) => sl<ExitBloc>(),
          child: ExitQrScreen(
            orderId: orderId,
            token: token,
            expiresAt: expiresAt,
          ),
        );
      },
    ),
    GoRoute(
      path: '/exit/success',
      builder: (context, state) => const ExitSuccessScreen(),
    ),
    GoRoute(
      path: '/exit/expired',
      builder: (context, state) => const ExitExpiredScreen(),
    ),
    GoRoute(
      path: '/exit/help-needed',
      builder: (context, state) => const ExitHelpNeededScreen(),
    ),
    GoRoute(path: '/exit', builder: (context, state) => const ExitScreen()),
    GoRoute(
      path: '/loyalty',
      builder: (context, state) => const LoyaltyHomeScreen(),
    ),
    GoRoute(
      path: '/loyalty/history',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<LoyaltyHistoryCubit>(),
        child: const LoyaltyHistoryScreen(),
      ),
    ),
    GoRoute(
      path: '/loyalty/tiers',
      builder: (context, state) => RepositoryProvider.value(
        value: sl<GetTiersInfoUseCase>(),
        child: const LoyaltyTiersInfoScreen(),
      ),
    ),
    GoRoute(
      path: '/loyalty/referral',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<ReferralCubit>(),
        child: const ReferralScreen(),
      ),
    ),
    GoRoute(path: '/orders', builder: (context, state) => const OrdersScreen()),
    GoRoute(path: '/notifications', builder: (context, state) => const NotificationsScreen()),
    GoRoute(path: '/profile', builder: (context, state) => const ProfileScreen()),
    GoRoute(
      path: '/profile/edit',
      builder: (context, state) {
        final extra = state.extra as Map<String, dynamic>?;
        return AccountEditScreen(
          currentName: extra?['name'] as String? ?? 'Rahul Mehta',
          currentPhone: extra?['phone'] as String? ?? '+91 98765 43210',
          currentEmail: extra?['email'] as String? ?? 'krishna@gmail.com',
          currentBirthday: extra?['birthday'] as String? ?? 'Mar 12, 1994',
          currentLanguage: extra?['language'] as String? ?? 'English',
        );
      },
    ),
    GoRoute(path: '/settings', builder: (context, state) => const SettingsScreen()),
    GoRoute(path: '/profile/data', builder: (context, state) => const ManageMyDataScreen()),
    GoRoute(
      path: '/profile/sessions',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<DeviceSessionsBloc>(),
        child: const DeviceSessionsScreen(),
      ),
    ),
    GoRoute(
      path: '/profile/membership',
      builder: (context, state) => BlocProvider(
        create: (_) => sl<MembershipCubit>(),
        child: const MembershipScreen(),
      ),
    ),
  ],
);

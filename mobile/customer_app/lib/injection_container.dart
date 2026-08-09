import 'package:get_it/get_it.dart';
import 'package:zippyra_core/zippyra_core.dart';

import 'features/auth/data/datasources/auth_remote_data_source.dart';
import 'features/auth/data/repositories/auth_repository_impl.dart';
import 'features/auth/domain/repositories/auth_repository.dart';
import 'features/auth/domain/usecases/send_otp_use_case.dart';
import 'features/auth/domain/usecases/verify_otp_use_case.dart';
import 'features/auth/domain/usecases/sign_in_with_google_use_case.dart';
import 'features/auth/presentation/bloc/auth_bloc.dart';
import 'core/services/deep_link_service.dart';
import 'core/services/play_integrity_service.dart';
import 'core/services/root_detection_service.dart';
import 'core/router/app_router.dart';

import 'features/profile/data/datasources/device_sessions_remote_data_source.dart';
import 'features/profile/data/repositories/device_sessions_repository_impl.dart';
import 'features/profile/domain/repositories/device_sessions_repository.dart';
import 'features/profile/presentation/bloc/device_sessions_bloc.dart';

import 'features/loyalty/data/datasources/referral_remote_data_source.dart';
import 'features/loyalty/data/repositories/referral_repository_impl.dart';
import 'features/loyalty/domain/repositories/referral_repository.dart';
import 'features/loyalty/presentation/cubit/referral_cubit.dart';

import 'features/membership/data/datasources/membership_remote_data_source.dart';
import 'features/membership/data/repositories/membership_repository_impl.dart';
import 'features/membership/domain/repositories/membership_repository.dart';
import 'features/membership/presentation/cubit/membership_cubit.dart';

import 'features/feedback/data/feedback_service.dart';

import 'features/notifications/data/datasources/notifications_remote_data_source.dart';
import 'features/notifications/data/repositories/notifications_repository_impl.dart';
import 'features/notifications/data/device_token_registrar.dart';
import 'features/notifications/domain/repositories/notifications_repository.dart';
import 'features/notifications/domain/usecases/get_preferences_use_case.dart';
import 'features/notifications/domain/usecases/update_preference_use_case.dart';
import 'features/notifications/domain/usecases/get_inbox_use_case.dart';
import 'features/notifications/domain/usecases/mark_read_use_case.dart';
import 'features/notifications/domain/usecases/get_unread_count_use_case.dart';
import 'features/notifications/domain/usecases/register_device_token_use_case.dart';
import 'features/notifications/domain/usecases/unregister_device_token_use_case.dart';
import 'features/notifications/presentation/bloc/notifications_bloc.dart';

import 'features/store_session/data/datasources/store_remote_data_source.dart';
import 'features/store_session/data/repositories/store_session_repository_impl.dart';
import 'features/store_session/domain/repositories/store_session_repository.dart';
import 'features/store_session/domain/usecases/get_nearby_stores_use_case.dart';
import 'features/store_session/domain/usecases/bind_store_use_case.dart';
import 'features/store_session/domain/usecases/unbind_store_use_case.dart';
import 'features/store_session/domain/usecases/restore_session_use_case.dart';
import 'features/store_session/presentation/bloc/store_session_bloc.dart';

import 'features/catalog/data/datasources/catalog_local_data_source.dart';
import 'features/catalog/data/datasources/catalog_remote_data_source.dart';
import 'features/catalog/data/repositories/catalog_repository_impl.dart';
import 'features/catalog/domain/repositories/catalog_repository.dart';
import 'features/catalog/domain/usecases/get_categories_use_case.dart';
import 'features/catalog/domain/usecases/get_product_by_barcode_use_case.dart';
import 'features/catalog/domain/usecases/search_products_use_case.dart';
import 'features/catalog/domain/usecases/sync_catalog_use_case.dart';
import 'features/catalog/presentation/bloc/catalog_bloc.dart';

import 'features/cart/data/datasources/cart_remote_data_source.dart';
import 'features/cart/data/repositories/cart_repository_impl.dart';
import 'features/cart/domain/repositories/cart_repository.dart';
import 'features/cart/domain/usecases/apply_coupon_use_case.dart';
import 'features/cart/domain/usecases/clear_cart_use_case.dart';
import 'features/cart/domain/usecases/get_cart_use_case.dart';
import 'features/cart/domain/usecases/init_checkout_use_case.dart';
import 'features/cart/domain/usecases/remove_coupon_use_case.dart';
import 'features/cart/domain/usecases/remove_item_use_case.dart';
import 'features/cart/domain/usecases/scan_item_use_case.dart';
import 'features/cart/domain/usecases/update_quantity_use_case.dart';
import 'features/cart/presentation/bloc/cart_bloc.dart';

import 'features/payment/data/datasources/payment_remote_data_source.dart';
import 'features/payment/data/razorpay_service.dart';
import 'features/payment/data/repositories/payment_repository_impl.dart';
import 'features/payment/domain/repositories/payment_repository.dart';
import 'features/payment/domain/usecases/check_pending_payment_use_case.dart';
import 'features/payment/domain/usecases/estimate_split_use_case.dart';
import 'features/payment/domain/usecases/get_payment_status_use_case.dart';
import 'features/payment/domain/usecases/initiate_payment_use_case.dart';
import 'features/payment/presentation/bloc/payment_bloc.dart';

import 'features/orders/data/datasources/orders_remote_data_source.dart';
import 'features/orders/data/repositories/orders_repository_impl.dart';
import 'features/orders/domain/repositories/orders_repository.dart';
import 'features/orders/domain/usecases/get_exit_token_use_case.dart';
import 'features/orders/domain/usecases/get_order_detail_use_case.dart';
import 'features/orders/domain/usecases/get_order_history_use_case.dart';
import 'features/orders/domain/usecases/request_return_use_case.dart';
import 'features/orders/presentation/bloc/order_detail_bloc.dart';
import 'features/orders/presentation/bloc/order_history_bloc.dart';
import 'features/orders/presentation/cubit/order_exit_cubit.dart';

import 'features/exit/data/datasources/exit_remote_data_source.dart';
import 'features/exit/data/repositories/exit_repository_impl.dart';
import 'features/exit/domain/repositories/exit_repository.dart';
import 'features/exit/domain/usecases/poll_exit_status_use_case.dart';
import 'features/exit/presentation/bloc/exit_bloc.dart';

import 'features/loyalty/data/datasources/loyalty_remote_data_source.dart';
import 'features/loyalty/data/repositories/loyalty_repository_impl.dart';
import 'features/loyalty/domain/repositories/loyalty_repository.dart';
import 'features/loyalty/domain/usecases/get_loyalty_balance_use_case.dart';
import 'features/loyalty/domain/usecases/get_loyalty_history_use_case.dart';
import 'features/loyalty/domain/usecases/get_tiers_info_use_case.dart';
import 'features/loyalty/presentation/bloc/loyalty_bloc.dart';
import 'features/loyalty/presentation/cubit/loyalty_history_cubit.dart';

import 'features/home/data/datasources/home_remote_data_source.dart';
import 'features/home/data/repositories/home_repository_impl.dart';
import 'features/home/domain/repositories/home_repository.dart';
import 'features/home/domain/usecases/get_home_banners_use_case.dart';
import 'features/home/presentation/bloc/home_bloc.dart';

final sl = GetIt.instance;

Future<void> initServiceLocator() async {
  // Core Lazy Singletons
  sl.registerLazySingleton<ApiClient>(() => ApiClient());
  sl.registerLazySingleton<SecureStorage>(() => SecureStorage());

  // Auth Data Layer
  sl.registerLazySingleton<AuthRemoteDataSource>(
    () => AuthRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<AuthRepository>(
    () => AuthRepositoryImpl(
      remoteDataSource: sl(),
      secureStorage: sl(),
    ),
  );

  // Auth Domain Layer (UseCases)
  sl.registerLazySingleton<SendOtpUseCase>(() => SendOtpUseCase(sl()));
  sl.registerLazySingleton<VerifyOtpUseCase>(() => VerifyOtpUseCase(sl()));
  sl.registerLazySingleton<SignInWithGoogleUseCase>(() => SignInWithGoogleUseCase(sl()));

  // Auth Presentation Layer (Bloc)
  sl.registerFactory<AuthBloc>(
    () => AuthBloc(
      sendOtpUseCase: sl(),
      verifyOtpUseCase: sl(),
      signInWithGoogleUseCase: sl(),
    ),
  );

  // Store Session Data Layer
  sl.registerLazySingleton<StoreRemoteDataSource>(
    () => StoreRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<StoreSessionRepository>(
    () => StoreSessionRepositoryImpl(
      remoteDataSource: sl(),
      secureStorage: sl(),
    ),
  );

  // Store Session Domain Layer (UseCases)
  sl.registerLazySingleton<GetNearbyStoresUseCase>(() => GetNearbyStoresUseCase(sl()));
  sl.registerLazySingleton<BindStoreUseCase>(() => BindStoreUseCase(sl()));
  sl.registerLazySingleton<UnbindStoreUseCase>(() => UnbindStoreUseCase(sl()));
  sl.registerLazySingleton<RestoreSessionUseCase>(() => RestoreSessionUseCase(sl()));

  // Store Session Presentation Layer (Bloc)
  sl.registerLazySingleton<StoreSessionBloc>(
    () => StoreSessionBloc(
      getNearbyStoresUseCase: sl(),
      bindStoreUseCase: sl(),
      unbindStoreUseCase: sl(),
      restoreSessionUseCase: sl(),
      secureStorage: sl(),
    ),
  );

  // Catalog Data Layer
  sl.registerLazySingleton<CatalogDatabase>(() => CatalogDatabase.instance);
  sl.registerLazySingleton<CatalogLocalDataSource>(
    () => CatalogLocalDataSourceImpl(database: sl()),
  );
  sl.registerLazySingleton<CatalogRemoteDataSource>(
    () => CatalogRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<CatalogSyncEngine>(
    () => CatalogSyncEngine(remoteDataSource: sl(), database: sl()),
  );
  sl.registerLazySingleton<CatalogRepository>(
    () => CatalogRepositoryImpl(
      localDataSource: sl(),
      remoteDataSource: sl(),
      syncEngine: sl(),
    ),
  );

  // Catalog Domain Layer (UseCases)
  sl.registerLazySingleton<GetProductByBarcodeUseCase>(() => GetProductByBarcodeUseCase(sl()));
  sl.registerLazySingleton<SearchProductsUseCase>(() => SearchProductsUseCase(sl()));
  sl.registerLazySingleton<GetCategoriesUseCase>(() => GetCategoriesUseCase(sl()));
  sl.registerLazySingleton<SyncCatalogUseCase>(() => SyncCatalogUseCase(sl()));

  // Catalog Presentation Layer (Bloc)
  sl.registerLazySingleton<CatalogBloc>(
    () => CatalogBloc(
      getProductByBarcodeUseCase: sl(),
      searchProductsUseCase: sl(),
      getCategoriesUseCase: sl(),
      syncCatalogUseCase: sl(),
    ),
  );

  // Cart Data Layer
  sl.registerLazySingleton<CartRemoteDataSource>(
    () => CartRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<CartRepository>(
    () => CartRepositoryImpl(remoteDataSource: sl()),
  );

  // Cart Domain Layer (UseCases)
  sl.registerLazySingleton<GetCartUseCase>(() => GetCartUseCase(sl()));
  sl.registerLazySingleton<ScanItemUseCase>(() => ScanItemUseCase(sl()));
  sl.registerLazySingleton<UpdateQuantityUseCase>(() => UpdateQuantityUseCase(sl()));
  sl.registerLazySingleton<RemoveItemUseCase>(() => RemoveItemUseCase(sl()));
  sl.registerLazySingleton<ClearCartUseCase>(() => ClearCartUseCase(sl()));
  sl.registerLazySingleton<ApplyCouponUseCase>(() => ApplyCouponUseCase(sl()));
  sl.registerLazySingleton<RemoveCouponUseCase>(() => RemoveCouponUseCase(sl()));
  sl.registerLazySingleton<InitCheckoutUseCase>(() => InitCheckoutUseCase(sl()));

  // Cart Presentation Layer (Bloc) - Scoped Lazy Singleton
  sl.registerLazySingleton<CartBloc>(
    () => CartBloc(
      getCartUseCase: sl(),
      scanItemUseCase: sl(),
      updateQuantityUseCase: sl(),
      removeItemUseCase: sl(),
      clearCartUseCase: sl(),
      applyCouponUseCase: sl(),
      removeCouponUseCase: sl(),
      initCheckoutUseCase: sl(),
    ),
  );

  // Security & Integrity Services
  sl.registerLazySingleton<PlayIntegrityService>(() => PlayIntegrityServiceImpl());
  sl.registerLazySingleton<CustomerRootDetectionService>(() => CustomerRootDetectionService());

  // Payment Layer
  sl.registerLazySingleton<RazorpayService>(() => RazorpayService());
  sl.registerLazySingleton<PaymentRemoteDataSource>(
    () => PaymentRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<PaymentRepository>(
    () => PaymentRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<EstimateSplitUseCase>(() => EstimateSplitUseCase(sl()));
  sl.registerLazySingleton<InitiatePaymentUseCase>(() => InitiatePaymentUseCase(sl()));
  sl.registerLazySingleton<GetPaymentStatusUseCase>(() => GetPaymentStatusUseCase(sl()));
  sl.registerLazySingleton<CheckPendingPaymentUseCase>(() => CheckPendingPaymentUseCase(sl()));

  sl.registerFactory<PaymentBloc>(
    () => PaymentBloc(
      estimateSplitUseCase: sl(),
      initiatePaymentUseCase: sl(),
      getPaymentStatusUseCase: sl(),
      checkPendingPaymentUseCase: sl(),
      playIntegrityService: sl(),
      razorpayService: sl(),
    ),
  );

  // Orders Layer
  sl.registerLazySingleton<OrdersRemoteDataSource>(
    () => OrdersRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<OrdersRepository>(
    () => OrdersRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<GetOrderHistoryUseCase>(() => GetOrderHistoryUseCase(sl()));
  sl.registerLazySingleton<GetOrderDetailUseCase>(() => GetOrderDetailUseCase(sl()));
  sl.registerLazySingleton<RequestReturnUseCase>(() => RequestReturnUseCase(sl()));
  sl.registerLazySingleton<GetExitTokenUseCase>(() => GetExitTokenUseCase(sl()));

  sl.registerFactory<OrderHistoryBloc>(
    () => OrderHistoryBloc(getOrderHistoryUseCase: sl()),
  );
  sl.registerFactory<OrderDetailBloc>(
    () => OrderDetailBloc(
      getOrderDetailUseCase: sl(),
      requestReturnUseCase: sl(),
    ),
  );
  sl.registerFactory<OrderExitCubit>(
    () => OrderExitCubit(getExitTokenUseCase: sl()),
  );

  // Exit Layer
  sl.registerLazySingleton<ExitRemoteDataSource>(
    () => ExitRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<ExitRepository>(
    () => ExitRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<PollExitStatusUseCase>(() => PollExitStatusUseCase(sl()));

  sl.registerFactory<ExitBloc>(
    () => ExitBloc(pollExitStatusUseCase: sl()),
  );

  // Loyalty Layer
  sl.registerLazySingleton<LoyaltyRemoteDataSource>(
    () => LoyaltyRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<LoyaltyRepository>(
    () => LoyaltyRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<GetLoyaltyBalanceUseCase>(() => GetLoyaltyBalanceUseCase(sl()));
  sl.registerLazySingleton<GetLoyaltyHistoryUseCase>(() => GetLoyaltyHistoryUseCase(sl()));
  sl.registerLazySingleton<GetTiersInfoUseCase>(() => GetTiersInfoUseCase(sl()));

  sl.registerLazySingleton<LoyaltyBloc>(
    () => LoyaltyBloc(getLoyaltyBalanceUseCase: sl()),
  );
  sl.registerFactory<LoyaltyHistoryCubit>(
    () => LoyaltyHistoryCubit(getLoyaltyHistoryUseCase: sl()),
  );

  // Deep Link Service
  sl.registerLazySingleton<DeepLinkService>(
    () => DeepLinkService(router: appRouter),
  );

  // Notifications Layer
  sl.registerLazySingleton<NotificationsRemoteDataSource>(
    () => NotificationsRemoteDataSourceImpl(client: sl()),
  );
  sl.registerLazySingleton<NotificationsRepository>(
    () => NotificationsRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<GetPreferencesUseCase>(() => GetPreferencesUseCase(sl()));
  sl.registerLazySingleton<UpdatePreferenceUseCase>(() => UpdatePreferenceUseCase(sl()));
  sl.registerLazySingleton<GetInboxUseCase>(() => GetInboxUseCase(sl()));
  sl.registerLazySingleton<MarkReadUseCase>(() => MarkReadUseCase(sl()));
  sl.registerLazySingleton<GetUnreadCountUseCase>(() => GetUnreadCountUseCase(sl()));
  sl.registerLazySingleton<RegisterDeviceTokenUseCase>(() => RegisterDeviceTokenUseCase(sl()));
  sl.registerLazySingleton<UnregisterDeviceTokenUseCase>(() => UnregisterDeviceTokenUseCase(sl()));

  sl.registerLazySingleton<DeviceTokenRegistrar>(
    () => DeviceTokenRegistrarImpl(
      registerDeviceTokenUseCase: sl(),
      unregisterDeviceTokenUseCase: sl(),
      secureStorage: sl(),
    ),
  );

  sl.registerLazySingleton<NotificationsBloc>(
    () => NotificationsBloc(
      getPreferencesUseCase: sl(),
      updatePreferenceUseCase: sl(),
      getInboxUseCase: sl(),
      markReadUseCase: sl(),
      getUnreadCountUseCase: sl(),
    ),
  );

  // Home Layer
  sl.registerLazySingleton<HomeRemoteDataSource>(
    () => HomeRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<HomeRepository>(
    () => HomeRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<GetHomeBannersUseCase>(
    () => GetHomeBannersUseCase(sl()),
  );
  sl.registerFactory<HomeBloc>(
    () => HomeBloc(
      getHomeBannersUseCase: sl(),
      getNearbyStoresUseCase: sl(),
    ),
  );

  // Device Sessions Layer
  sl.registerLazySingleton<DeviceSessionsRemoteDataSource>(
    () => DeviceSessionsRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<DeviceSessionsRepository>(
    () => DeviceSessionsRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerFactory<DeviceSessionsBloc>(
    () => DeviceSessionsBloc(repository: sl()),
  );

  // Referral Layer
  sl.registerLazySingleton<ReferralRemoteDataSource>(
    () => ReferralRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<ReferralRepository>(
    () => ReferralRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerFactory<ReferralCubit>(
    () => ReferralCubit(repository: sl()),
  );

  // Membership Layer
  sl.registerLazySingleton<MembershipRemoteDataSource>(
    () => MembershipRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<MembershipRepository>(
    () => MembershipRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerFactory<MembershipCubit>(
    () => MembershipCubit(repository: sl()),
  );

  // Feedback Service
  sl.registerLazySingleton<FeedbackService>(
    () => FeedbackService(
      apiClient: sl(),
      secureStorage: sl(),
    ),
  );
}

import 'package:get_it/get_it.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'core/services/offline_queue_service.dart';
import 'core/services/mqtt_service.dart';
import 'core/services/rfid_service.dart';
import 'features/auth/data/datasources/auth_remote_data_source.dart';
import 'features/auth/data/repositories/auth_repository_impl.dart';
import 'features/auth/domain/repositories/auth_repository.dart';
import 'features/auth/presentation/bloc/auth_bloc.dart';
import 'features/shift/data/datasources/shift_remote_data_source.dart';
import 'features/shift/data/repositories/shift_repository_impl.dart';
import 'features/shift/domain/repositories/shift_repository.dart';
import 'features/shift/presentation/bloc/shift_bloc.dart';
import 'features/inventory/data/datasources/inventory_remote_data_source.dart';
import 'features/inventory/data/repositories/inventory_repository_impl.dart';
import 'features/inventory/domain/repositories/inventory_repository.dart';
import 'features/inventory/presentation/bloc/low_stock_bloc.dart';
import 'features/inventory/presentation/bloc/stock_count_bloc.dart';
import 'features/inventory/presentation/bloc/grn_bloc.dart';
import 'features/devices/data/repositories/devices_repository.dart';
import 'features/devices/presentation/bloc/devices_bloc.dart';

import 'features/device_pairing/data/datasources/device_pairing_remote_data_source.dart';
import 'features/device_pairing/data/repositories/device_pairing_repository_impl.dart';
import 'features/device_pairing/domain/repositories/device_pairing_repository.dart';
import 'features/device_pairing/domain/usecases/pair_device_use_case.dart';
import 'features/device_pairing/presentation/bloc/device_pairing_bloc.dart';

import 'features/pos_assist/presentation/bloc/price_check_bloc.dart';
import 'features/customer_assist/presentation/bloc/customer_lookup_bloc.dart';

import 'core/services/root_detection_service.dart';

final sl = GetIt.instance;

Future<void> initServiceLocator() async {
  // Core Infrastructure
  sl.registerLazySingleton<FlutterSecureStorage>(() => const FlutterSecureStorage());
  sl.registerLazySingleton<SecureStorage>(() => SecureStorage());
  sl.registerLazySingleton<ApiClient>(() => ApiClient());
  sl.registerLazySingleton<OfflineQueueService>(() => OfflineQueueService());
  sl.registerLazySingleton<StaffRootDetectionService>(() => StaffRootDetectionService());
  sl.registerLazySingleton<CatalogDatabase>(() => CatalogDatabase.instance);
  sl.registerLazySingleton<MqttService>(() => MqttService(secureStorage: sl()));
  sl.registerLazySingleton<RfidService>(() => RfidServiceImpl());

  // Devices Layer
  sl.registerLazySingleton<DevicesRepository>(() => DevicesRepositoryImpl());
  sl.registerFactory<DevicesBloc>(
    () => DevicesBloc(repository: sl(), mqttService: sl()),
  );

  // Device Pairing Layer
  sl.registerLazySingleton<DevicePairingRemoteDataSource>(
    () => DevicePairingRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<DevicePairingRepository>(
    () => DevicePairingRepositoryImpl(remoteDataSource: sl(), secureStorage: sl()),
  );
  sl.registerLazySingleton<PairDeviceUseCase>(() => PairDeviceUseCase(sl()));
  sl.registerLazySingleton<CheckDevicePairedUseCase>(() => CheckDevicePairedUseCase(sl()));
  sl.registerLazySingleton<ClearDevicePairingUseCase>(() => ClearDevicePairingUseCase(sl()));

  sl.registerLazySingleton<DevicePairingBloc>(
    () => DevicePairingBloc(
      pairDeviceUseCase: sl(),
      checkDevicePairedUseCase: sl(),
      clearDevicePairingUseCase: sl(),
    ),
  );

  // Auth Layer
  sl.registerLazySingleton<AuthRemoteDataSource>(
    () => AuthRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<AuthRepository>(
    () => AuthRepositoryImpl(remoteDataSource: sl(), secureStorage: sl()),
  );
  sl.registerLazySingleton<AuthBloc>(
    () => AuthBloc(authRepository: sl()),
  );

  // Shift Layer
  sl.registerLazySingleton<ShiftRemoteDataSource>(
    () => ShiftRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<ShiftRepository>(
    () => ShiftRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerLazySingleton<ShiftBloc>(
    () => ShiftBloc(shiftRepository: sl()),
  );

  // POS Assist & Customer Assist
  sl.registerFactory<PriceCheckBloc>(
    () => PriceCheckBloc(catalogDatabase: sl(), apiClient: sl()),
  );
  sl.registerFactory<CustomerLookupBloc>(
    () => CustomerLookupBloc(apiClient: sl()),
  );

  // Inventory Layer
  sl.registerLazySingleton<InventoryRemoteDataSource>(
    () => InventoryRemoteDataSourceImpl(apiClient: sl()),
  );
  sl.registerLazySingleton<InventoryRepository>(
    () => InventoryRepositoryImpl(remoteDataSource: sl()),
  );
  sl.registerFactory<LowStockBloc>(
    () => LowStockBloc(repository: sl()),
  );
  sl.registerFactory<StockCountBloc>(
    () => StockCountBloc(repository: sl(), offlineQueueService: sl()),
  );
  sl.registerFactory<GrnBloc>(
    () => GrnBloc(repository: sl()),
  );
}

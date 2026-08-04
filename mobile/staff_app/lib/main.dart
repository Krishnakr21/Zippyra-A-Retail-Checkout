import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'core/router/app_router.dart';
import 'features/auth/presentation/bloc/auth_bloc.dart';
import 'features/shift/presentation/bloc/shift_bloc.dart';
import 'features/inventory/presentation/bloc/low_stock_bloc.dart';
import 'features/inventory/presentation/bloc/stock_count_bloc.dart';
import 'features/inventory/presentation/bloc/grn_bloc.dart';
import 'injection_container.dart';
import 'core/services/root_detection_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initServiceLocator();
  // Non-blocking root check (log and alert only)
  sl<StaffRootDetectionService>().checkRootStatus();

  runApp(const StaffApp());
}

class StaffApp extends StatelessWidget {
  const StaffApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider<AuthBloc>(
          create: (_) => sl<AuthBloc>()..add(AuthRestoreSessionRequested()),
        ),
        BlocProvider<ShiftBloc>(
          create: (_) => sl<ShiftBloc>(),
        ),
        BlocProvider<LowStockBloc>(
          create: (_) => sl<LowStockBloc>(),
        ),
        BlocProvider<StockCountBloc>(
          create: (_) => sl<StockCountBloc>(),
        ),
        BlocProvider<GrnBloc>(
          create: (_) => sl<GrnBloc>(),
        ),
      ],
      child: MaterialApp.router(
        title: 'Zippyra Staff',
        theme: ZippyraTheme.lightTheme,
        routerConfig: staffAppRouter,
        localizationsDelegates: const [
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        supportedLocales: const [
          Locale('en', ''),
          Locale('hi', ''),
        ],
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';

import 'core/router/app_router.dart';
import 'core/services/root_detection_service.dart';
import 'features/cart/presentation/bloc/cart_bloc.dart';
import 'injection_container.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initServiceLocator();
  // Non-blocking root check (after session restore)
  sl<CustomerRootDetectionService>().checkRootStatus();

  runApp(const ZippyraCustomerApp());
}

class ZippyraCustomerApp extends StatelessWidget {
  const ZippyraCustomerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiBlocProvider(
      providers: [
        BlocProvider<CartBloc>(
          create: (_) => sl<CartBloc>()..add(const CartRefreshRequested('store-1')),
        ),
      ],
      child: MaterialApp.router(
        title: 'Zippyra',
        theme: ZippyraTheme.lightTheme,
        routerConfig: appRouter,
        debugShowCheckedModeBanner: false,
      ),
    );
  }
}

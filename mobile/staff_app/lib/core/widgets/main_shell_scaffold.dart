import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:zippyra_core/zippyra_core.dart';
import '../../features/auth/presentation/bloc/auth_bloc.dart';
import '../../features/shift/presentation/screens/shift_home_screen.dart';
import '../../features/inventory/presentation/screens/low_stock_screen.dart';
import '../../features/devices/presentation/screens/devices_screen.dart';
import '../../features/pos_assist/presentation/screens/cash_payment_screen.dart';
import '../../features/profile/presentation/screens/profile_screen.dart';

class MainShellScaffold extends StatefulWidget {
  final int initialIndex;

  const MainShellScaffold({super.key, this.initialIndex = 0});

  @override
  State<MainShellScaffold> createState() => _MainShellScaffoldState();
}

class _MainShellScaffoldState extends State<MainShellScaffold> {
  late int _currentIndex;

  static const List<Widget> _allTabs = [
    ShiftHomeScreen(),
    LowStockScreen(),
    DevicesScreen(),
    CashPaymentScreen(),
    ProfileScreen(),
  ];

  @override
  void initState() {
    super.initState();
    _currentIndex = widget.initialIndex;
  }

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    String role = 'MANAGER';
    if (authState is AuthAuthenticated) {
      role = authState.role.toUpperCase();
    }

    // Role-based tab filtering: SECURITY role hides Inventory tab
    final showInventory = role != 'SECURITY';

    final items = [
      const BottomNavigationBarItem(
        icon: Icon(Icons.home_outlined),
        activeIcon: Icon(Icons.home),
        label: 'Shift',
      ),
      if (showInventory)
        const BottomNavigationBarItem(
          icon: Icon(Icons.inventory_2_outlined),
          activeIcon: Icon(Icons.inventory_2),
          label: 'Inventory',
        ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.sensors_outlined),
        activeIcon: Icon(Icons.sensors),
        label: 'Devices',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.point_of_sale_outlined),
        activeIcon: Icon(Icons.point_of_sale),
        label: 'POS Assist',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.person_outline),
        activeIcon: Icon(Icons.person),
        label: 'Profile',
      ),
    ];

    return Scaffold(
      body: IndexedStack(
        index: _currentIndex,
        children: _allTabs,
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex.clamp(0, items.length - 1),
        onTap: (index) {
          setState(() {
            _currentIndex = index;
          });
        },
        type: BottomNavigationBarType.fixed,
        selectedItemColor: ZippyraColors.primary,
        unselectedItemColor: Colors.grey,
        items: items,
      ),
    );
  }
}

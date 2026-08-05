import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/widgets/role_guard.dart';
import 'package:staff_app/features/auth/domain/entities/staff_session.dart';

import 'package:staff_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:staff_app/features/auth/presentation/bloc/auth_bloc.dart';

class DummyAuthRepo implements AuthRepository {
  @override
  Future<void> logout() async {}

  @override
  Future<StaffSession?> restoreSession() async => null;

  @override
  Future<void> sendOtp(String phone) async {}

  @override
  Future<void> setPin(String pin) async {}

  @override
  Future<StaffSession> verifyOtp(String phone, String otp) async => const StaffSession(
        token: 'token',
        staffId: '1',
        role: 'CASHIER',
        storeId: 'store-1',
        storeName: 'Store 1',
      );

  @override
  Future<StaffSession> loginWithPin(String phone, String pin) async => const StaffSession(
        token: 'token',
        staffId: '1',
        role: 'CASHIER',
        storeId: 'store-1',
        storeName: 'Store 1',
      );
}

void main() {
  testWidgets('RoleGuard renders child when role is authorized', (tester) async {
    final repo = DummyAuthRepo();
    final authBloc = AuthBloc(authRepository: repo);
    authBloc.emit(const AuthAuthenticated(
      staffId: '101',
      role: 'CASHIER',
      storeId: 'store-1',
      storeName: 'Downtown Store',
      session: StaffSession(token: 't', staffId: '101', role: 'CASHIER', storeId: 's1', storeName: 's'),
    ));

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<AuthBloc>.value(
          value: authBloc,
          child: const RoleGuard(
            allowedRoles: ['CASHIER', 'MANAGER'],
            child: Text('Protected Content'),
          ),
        ),
      ),
    );

    expect(find.text('Protected Content'), findsOneWidget);
  });
}

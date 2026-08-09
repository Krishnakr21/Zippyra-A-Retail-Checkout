import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:zippyra_core/theme/zippyra_theme.dart';
import 'package:zippyra_core/widgets/zbutton.dart';
import 'package:zippyra_core/widgets/zinput.dart';

class PhoneEntryScreen extends StatefulWidget {
  const PhoneEntryScreen({super.key});

  @override
  State<PhoneEntryScreen> createState() => _PhoneEntryScreenState();
}

class _PhoneEntryScreenState extends State<PhoneEntryScreen> {
  final TextEditingController _phoneController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Login or Signup')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Enter Mobile Number',
              style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 6),
            const Text(
              'We will send a 6-digit OTP to verify your account',
              style: TextStyle(color: ZippyraColors.textSecondary),
            ),
            const SizedBox(height: 32),
            ZInput(
              label: 'Mobile Number',
              hint: '+91 98765 43210',
              controller: _phoneController,
              keyboardType: TextInputType.phone,
              prefixIcon: const Icon(Icons.phone, color: ZippyraColors.primaryBlue),
            ),
            const Spacer(),
            ZButton(
              label: 'Get OTP',
              onPressed: () => context.push('/auth/otp'),
            ),
          ],
        ),
      ),
    );
  }
}

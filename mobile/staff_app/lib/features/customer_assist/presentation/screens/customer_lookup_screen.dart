import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../bloc/customer_lookup_bloc.dart';
import '../bloc/customer_lookup_event.dart';
import '../bloc/customer_lookup_state.dart';

class CustomerLookupScreen extends StatefulWidget {
  final String storeId;

  const CustomerLookupScreen({Key? key, this.storeId = 'store-demo-1'}) : super(key: key);

  @override
  State<CustomerLookupScreen> createState() => _CustomerLookupScreenState();
}

class _CustomerLookupScreenState extends State<CustomerLookupScreen> {
  final TextEditingController _last4Controller = TextEditingController();

  @override
  void dispose() {
    _last4Controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Customer Assist'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue.shade50,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.blue.shade200),
              ),
              child: const Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.blue),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      'This lookup is scoped to active customer sessions and orders at your store within the last 2 hours for support purposes.',
                      style: TextStyle(fontSize: 13, color: Colors.black87),
                      key: Key('customer_assist_scope_note'),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
            const Text(
              'Enter Last 4 Digits of Phone Number',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _last4Controller,
                    keyboardType: TextInputType.number,
                    maxLength: 4,
                    style: const TextStyle(fontSize: 20, letterSpacing: 4, fontWeight: FontWeight.bold),
                    decoration: InputDecoration(
                      hintText: '1234',
                      counterText: '',
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                ElevatedButton(
                  key: const Key('btn_find_customer'),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
                    backgroundColor: Colors.indigo,
                  ),
                  onPressed: () {
                    context.read<CustomerLookupBloc>().add(
                          LookupRequested(
                            storeId: widget.storeId,
                            phoneLast4: _last4Controller.text,
                          ),
                        );
                  },
                  child: const Text('Find', style: TextStyle(color: Colors.white, fontSize: 16)),
                ),
              ],
            ),
            const SizedBox(height: 24),
            const Divider(),
            const SizedBox(height: 16),
            BlocBuilder<CustomerLookupBloc, CustomerLookupState>(
              builder: (context, state) {
                if (state is CustomerLookupSearching) {
                  return const Center(child: CircularProgressIndicator());
                }

                if (state is CustomerLookupFailed) {
                  return Center(
                    child: Text(
                      state.message,
                      style: const TextStyle(color: Colors.red, fontSize: 15),
                    ),
                  );
                }

                if (state is NoMatch) {
                  return Center(
                    child: Column(
                      children: [
                        const Icon(Icons.person_search_outlined, size: 64, color: Colors.grey),
                        const SizedBox(height: 12),
                        Text(
                          'No active customer found ending in "${state.phoneLast4}"',
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                        ),
                        const SizedBox(height: 4),
                        const Text(
                          'Only sessions & orders from the last 2 hours at this store are shown.',
                          style: TextStyle(fontSize: 12, color: Colors.grey),
                        ),
                      ],
                    ),
                  );
                }

                if (state is MultipleMatches) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Multiple Matches (${state.candidates.length}) - Select Customer:',
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.indigo),
                      ),
                      const SizedBox(height: 12),
                      ListView.separated(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        itemCount: state.candidates.length,
                        separatorBuilder: (_, __) => const Divider(),
                        itemBuilder: (context, index) {
                          final candidate = state.candidates[index];
                          return ListTile(
                            leading: const CircleAvatar(child: Icon(Icons.person)),
                            title: Text(candidate.firstName, style: const TextStyle(fontWeight: FontWeight.bold)),
                            subtitle: Text('${candidate.phoneMasked} • Order: ${candidate.activeOrderId ?? 'None'}'),
                            onTap: () {
                              context.read<CustomerLookupBloc>().emit(SingleMatch(candidate));
                            },
                          );
                        },
                      ),
                    ],
                  );
                }

                if (state is SingleMatch) {
                  final c = state.customer;
                  return Card(
                    elevation: 2,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    child: Padding(
                      padding: const EdgeInsets.all(20.0),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              CircleAvatar(
                                backgroundColor: Colors.indigo.shade50,
                                child: const Icon(Icons.person, color: Colors.indigo),
                              ),
                              const SizedBox(width: 12),
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    c.firstName,
                                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                                  ),
                                  Text(
                                    c.phoneMasked,
                                    style: const TextStyle(color: Colors.grey),
                                  ),
                                ],
                              ),
                            ],
                          ),
                          const Divider(height: 24),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text('Store Session:'),
                              Chip(
                                label: Text(c.hasActiveSession ? 'ACTIVE' : 'INACTIVE'),
                                backgroundColor: c.hasActiveSession ? Colors.green.shade50 : Colors.grey.shade100,
                                labelStyle: TextStyle(
                                  color: c.hasActiveSession ? Colors.green.shade800 : Colors.grey,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                          if (c.activeOrderId != null) ...[
                            const SizedBox(height: 8),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                const Text('Order Status:'),
                                Text(
                                  c.activeOrderStatus ?? 'N/A',
                                  style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.indigo),
                                ),
                              ],
                            ),
                          ],
                        ],
                      ),
                    ),
                  );
                }

                return const Center(
                  child: Text(
                    'Enter phone last 4 digits to verify customer session/order status.',
                    style: TextStyle(color: Colors.grey),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

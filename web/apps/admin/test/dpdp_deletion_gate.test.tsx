import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import DPDPRequestsPage from '../app/dashboard/compliance/dpdp-requests/page';

describe('DPDP Deletion Action Dual-Gate Enforcement', () => {
  beforeEach(() => {
    global.fetch = jest.fn((url: string | URL | Request) => {
      const urlStr = url.toString();
      if (urlStr.includes('/v1/compliance/requests')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            requests: [
              {
                id: 'req-del-001',
                user_id: 'user-to-delete-123',
                request_type: 'DELETION',
                status: 'RECEIVED',
                sla_due_at: '2026-08-30T00:00:00Z',
                created_at: '2026-08-01T00:00:00Z',
              },
            ],
          }),
        });
      }
      if (urlStr.includes('/process-deletion')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ status: 'IN_PROGRESS' }),
        });
      }
      return Promise.resolve({ ok: true, json: async () => ({}) });
    }) as jest.Mock;
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('Requires BOTH ConfirmDialog AND StepUp re-authentication before process-deletion API fires', async () => {
    render(<DPDPRequestsPage />);

    // Wait for request queue to load
    await waitFor(() => {
      expect(screen.getByText('user-to-delete-123')).toBeInTheDocument();
    });

    const processBtn = screen.getByTestId('process-deletion-btn-req-del-001');

    // 1. Initial state: process-deletion API has NOT fired
    let deletionCalls = (global.fetch as jest.Mock).mock.calls.filter((c) => c[0].includes('/process-deletion'));
    expect(deletionCalls.length).toBe(0);

    // 2. Click "Process Deletion" -> Gate 1: ConfirmDialog opens
    fireEvent.click(processBtn);

    expect(screen.getByText('Permanent Irreversible Deletion Warning')).toBeInTheDocument();
    expect(screen.getByText(/This will permanently and irreversibly delete this user's data across all services/i)).toBeInTheDocument();

    // API still not fired
    deletionCalls = (global.fetch as jest.Mock).mock.calls.filter((c) => c[0].includes('/process-deletion'));
    expect(deletionCalls.length).toBe(0);

    // 3. Confirm Dialog -> Gate 2: StepUp modal opens
    const confirmBtn = screen.getByTestId('confirm-deletion-btn');
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(screen.getByText('Step-Up Re-Authentication')).toBeInTheDocument();
    });

    // API still not fired before password verification
    deletionCalls = (global.fetch as jest.Mock).mock.calls.filter((c) => c[0].includes('/process-deletion'));
    expect(deletionCalls.length).toBe(0);

    // 4. Enter Step-Up password and submit
    const passwordInput = screen.getByPlaceholderText('Enter admin password...');
    fireEvent.change(passwordInput, { target: { value: 'admin123' } });

    const stepUpSubmitBtn = screen.getByTestId('step-up-submit-btn');
    fireEvent.click(stepUpSubmitBtn);

    // 5. Now BOTH gates have passed -> process-deletion API call fires!
    await waitFor(() => {
      const finalCalls = (global.fetch as jest.Mock).mock.calls.filter((c) => c[0].includes('/process-deletion'));
      expect(finalCalls.length).toBe(1);
    });
  });
});

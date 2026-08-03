import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import HQERPPage from '../app/dashboard/erp/page';

describe('HQ ERP Page Role Gating & Read-Only Controls', () => {
  beforeEach(() => {
    // Mock global fetch for connection list
    global.fetch = jest.fn().mockImplementation((url: string) => {
      if (url.includes('/v1/integration/connections')) {
        return Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              connections: [
                {
                  id: 'conn-sap-1',
                  chain_id: 'chain-001',
                  erp_type: 'SAP',
                  integration_mode: 'DIRECT',
                  display_name: 'SAP Production Cloud',
                  enabled_outbound_events: ['order.completed'],
                  status: 'ACTIVE',
                  created_at: '2026-08-01T00:00:00Z',
                },
              ],
            }),
        });
      }
      return Promise.reject(new Error('Unknown URL'));
    });
  });

  test('FINANCE/OPERATIONS session sees no New Connection, Pause, or Rotate Secret buttons', async () => {
    render(<HQERPPage />);

    // Switch simulated role to FINANCE
    const roleSelect = screen.getByRole('combobox');
    const { fireEvent } = require('@testing-library/react');
    fireEvent.change(roleSelect, { target: { value: 'FINANCE' } });

    // Verify New Connection button is NOT rendered
    expect(screen.queryByTestId('new-connection-btn')).not.toBeInTheDocument();

    // Verify Pause / Rotate Secret buttons are NOT rendered on connection card
    expect(screen.queryByTestId('toggle-status-btn-conn-sap-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('rotate-secret-btn-conn-sap-1')).not.toBeInTheDocument();

    // Verify connection card and View Details link ARE rendered read-only
    expect(await screen.findByText('SAP Production Cloud')).toBeInTheDocument();
    expect(screen.getByTestId('view-detail-btn-conn-sap-1')).toBeInTheDocument();
  });
});

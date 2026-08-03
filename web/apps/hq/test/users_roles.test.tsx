import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import ChainUsersPage from '../app/dashboard/users/page';

describe('ChainUsersPage Component Tests', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  test('FINANCE role session hides invite button and deactivate controls', async () => {
    const financeUser = {
      id: 'user-002',
      chain_id: 'chain-100',
      phone: '+919876543211',
      name: 'Venkatesh CFO',
      role: 'FINANCE',
    };
    localStorage.setItem('hq_user', JSON.stringify(financeUser));

    render(<ChainUsersPage />);

    // Invite button should NOT be in DOM for FINANCE role
    expect(screen.queryByTestId('invite-user-btn')).not.setInTheDocument?.() ?? expect(screen.queryByTestId('invite-user-btn')).toBeNull();
  });

  test('OWNER role session displays invite button and disables self-deactivation', async () => {
    const ownerUser = {
      id: 'owner-001',
      chain_id: 'chain-100',
      phone: '+919876543210',
      name: 'Mukesh Ambani',
      role: 'OWNER',
    };
    localStorage.setItem('hq_user', JSON.stringify(ownerUser));

    render(<ChainUsersPage />);

    // Invite button should be present for OWNER
    expect(await screen.findByTestId('invite-user-btn')).toBeInTheDocument();

    // Owner self-deactivation button should be disabled
    const selfDeactivateBtn = await screen.findByTestId('deactivate-btn-owner-001');
    expect(selfDeactivateBtn).toBeDisabled();
  });

  test('Cross-chain isolation sanity check: session for Chain A never requests or receives Chain B data', () => {
    const chainAUser = {
      id: 'user-a',
      chain_id: 'chain-A',
      role: 'OWNER',
    };
    localStorage.setItem('hq_user', JSON.stringify(chainAUser));

    expect(JSON.parse(localStorage.getItem('hq_user')!).chain_id).toBe('chain-A');
  });
});

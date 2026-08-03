import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { OfferTable, OfferTableItem } from '../src/OfferTable';

describe('OfferTable Component Role Gating & Action Disabling', () => {
  const mockOffers: OfferTableItem[] = [
    {
      id: 'off-ss-1',
      chain_id: 'chain-1',
      store_id: 'store-1',
      type: 'PERCENT_OFF',
      applies_to: 'ALL',
      rule_config: { percent: 10 },
      min_cart_value_paise: 0,
      priority: 100,
      active_from: '2026-08-01T00:00:00Z',
      is_active: true,
      scope: 'STORE_SPECIFIC',
    },
    {
      id: 'off-cw-1',
      chain_id: 'chain-1',
      store_id: null,
      type: 'FLAT_OFF',
      applies_to: 'ALL',
      rule_config: { flat_paise: 500 },
      min_cart_value_paise: 0,
      priority: 50,
      active_from: '2026-08-01T00:00:00Z',
      is_active: true,
      scope: 'CHAIN_WIDE',
    },
  ];

  test('Disables Edit/Delete buttons on chain-wide rows for Retailer MANAGER role', () => {
    render(
      <OfferTable
        offers={mockOffers}
        showScope={true}
        userRole="MANAGER"
        onEdit={jest.fn()}
        onDelete={jest.fn()}
      />
    );

    // Store-specific row: Edit & Delete enabled
    const ssEditBtn = screen.getByTestId('edit-offer-btn-off-ss-1');
    const ssDeleteBtn = screen.getByTestId('delete-offer-btn-off-ss-1');
    expect(ssEditBtn).toBeEnabled();
    expect(ssDeleteBtn).toBeEnabled();

    // Chain-wide row: Edit & Delete DISABLED for MANAGER
    const cwEditBtn = screen.getByTestId('edit-offer-btn-off-cw-1');
    const cwDeleteBtn = screen.getByTestId('delete-offer-btn-off-cw-1');
    expect(cwEditBtn).toBeDisabled();
    expect(cwDeleteBtn).toBeDisabled();
    expect(cwEditBtn).toHaveAttribute('title', "Managed by your chain's HQ team.");
  });

  test('Disables Edit/Delete buttons on all rows for Chain HQ FINANCE role', () => {
    render(
      <OfferTable
        offers={mockOffers}
        showScope={true}
        userRole="FINANCE"
        onEdit={jest.fn()}
        onDelete={jest.fn()}
      />
    );

    const ssEditBtn = screen.getByTestId('edit-offer-btn-off-ss-1');
    const cwEditBtn = screen.getByTestId('edit-offer-btn-off-cw-1');

    expect(ssEditBtn).toBeDisabled();
    expect(cwEditBtn).toBeDisabled();
    expect(ssEditBtn).toHaveAttribute('title', 'Only OWNER role can modify offers.');
  });
});

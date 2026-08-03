import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { FunnelChart } from '../src/FunnelChart';

describe('FunnelChart Component', () => {
  test('Renders all 5 stages even when passed a response with zero counts for later stages', () => {
    const partialData = [
      { stage: 'SESSION_STARTED', sessionCount: 150, conversionFromPreviousPercent: 100 },
      { stage: 'CHECKOUT_INITIATED', sessionCount: 45, conversionFromPreviousPercent: 30 },
      // Later stages missing/zero from backend response
    ];

    render(<FunnelChart stages={partialData} />);

    expect(screen.getByTestId('funnel-chart')).toBeInTheDocument();

    // Verify all 5 stages are rendered
    expect(screen.getByTestId('funnel-stage-SESSION_STARTED')).toBeInTheDocument();
    expect(screen.getByTestId('funnel-stage-CHECKOUT_INITIATED')).toBeInTheDocument();
    expect(screen.getByTestId('funnel-stage-PAYMENT_CONFIRMED')).toBeInTheDocument();
    expect(screen.getByTestId('funnel-stage-ORDER_COMPLETED')).toBeInTheDocument();
    expect(screen.getByTestId('funnel-stage-EXIT_VALIDATED')).toBeInTheDocument();

    // Verify stage labels
    expect(screen.getByText('1. Session Started')).toBeInTheDocument();
    expect(screen.getByText('2. Checkout Initiated')).toBeInTheDocument();
    expect(screen.getByText('3. Payment Confirmed')).toBeInTheDocument();
    expect(screen.getByText('4. Order Completed')).toBeInTheDocument();
    expect(screen.getByText('5. Exit Validated')).toBeInTheDocument();
  });
});

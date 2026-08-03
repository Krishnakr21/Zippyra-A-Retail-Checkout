import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { PeakHoursHeatmap } from '../src/PeakHoursHeatmap';

describe('PeakHoursHeatmap Component', () => {
  test('Displays staffing caveat caption and renders hover tooltip', () => {
    const mockGrid = [
      { dayOfWeek: 1, hour: 14, avgTransactionsPerWeek: 85, recommendedStaff: 4 },
      { dayOfWeek: 5, hour: 18, avgTransactionsPerWeek: 120, recommendedStaff: 6 },
    ];

    render(<PeakHoursHeatmap grid={mockGrid} />);

    expect(screen.getByTestId('peak-hours-heatmap')).toBeInTheDocument();

    // Verify mandatory staffing formula caveat caption is present
    const caption = screen.getByTestId('staffing-caveat-caption');
    expect(caption).toBeInTheDocument();
    expect(caption).toHaveTextContent(
      'Staffing suggestion based on a simple throughput estimate — use judgment for final scheduling'
    );

    // Hover cell to test tooltip
    const cell = screen.getByTestId('heatmap-cell-1-14');
    fireEvent.mouseEnter(cell);

    expect(screen.getByTestId('heatmap-tooltip')).toBeInTheDocument();
    expect(screen.getByText(/Mon @ 14:00 - 15:00/i)).toBeInTheDocument();
  });
});

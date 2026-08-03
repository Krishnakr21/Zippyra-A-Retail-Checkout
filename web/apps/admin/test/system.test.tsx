import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import SystemOpsPage from '../app/dashboard/system/page';

// Mock useSystemOps hook
const mockGetDlqTopics = jest.fn();
const mockPeekDlqMessages = jest.fn();
const mockReplayDlqMessages = jest.fn();
const mockDiscardDlqMessages = jest.fn();
const mockGetFeatureFlags = jest.fn();
const mockCreateFeatureFlag = jest.fn();
const mockUpdateFeatureFlag = jest.fn();
const mockDeleteFeatureFlag = jest.fn();
const mockGetCircuitBreakerStatus = jest.fn();

jest.mock('@zippyra/hooks', () => ({
  useSystemOps: () => ({
    getDlqTopics: mockGetDlqTopics,
    peekDlqMessages: mockPeekDlqMessages,
    replayDlqMessages: mockReplayDlqMessages,
    discardDlqMessages: mockDiscardDlqMessages,
    getFeatureFlags: mockGetFeatureFlags,
    createFeatureFlag: mockCreateFeatureFlag,
    updateFeatureFlag: mockUpdateFeatureFlag,
    deleteFeatureFlag: mockDeleteFeatureFlag,
    getCircuitBreakerStatus: mockGetCircuitBreakerStatus,
  }),
}));

describe('SystemOpsPage Component Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetDlqTopics.mockResolvedValue({
      dlq_topics: [
        {
          topic_name: 'order.completed.dlq',
          message_count: 3,
          oldest_message_age_seconds: 300,
        },
      ],
    });
    mockPeekDlqMessages.mockResolvedValue({
      topic: 'order.completed.dlq',
      messages: [
        {
          topic: 'order.completed.dlq',
          offset: 101,
          key: 'ord-1',
          value: { order_id: 'ord-1', error: 'database error' },
          timestamp: '2026-08-01T00:00:00Z',
        },
      ],
      total: 1,
    });
    mockGetFeatureFlags.mockResolvedValue({
      feature_flags: [
        {
          flag_key: 'cart.dynamic_discounts',
          description: 'Dynamic cart discount engine',
          scope_type: 'GLOBAL',
          enabled_globally: true,
          updated_by: 'admin-1',
          updated_at: '2026-08-01T00:00:00Z',
          created_at: '2026-08-01T00:00:00Z',
        },
        {
          flag_key: 'warehouse.qc_required',
          description: 'Store specific QC requirement',
          scope_type: 'STORE',
          enabled_globally: false,
          enabled_scope_ids: ['store-mumbai-01'],
          updated_by: 'admin-1',
          updated_at: '2026-08-01T00:00:00Z',
          created_at: '2026-08-01T00:00:00Z',
        },
        {
          flag_key: 'catalog.new_search',
          description: 'Gradual search engine rollout',
          scope_type: 'USER_PERCENTAGE',
          enabled_globally: false,
          user_percentage: 50,
          updated_by: 'admin-1',
          updated_at: '2026-08-01T00:00:00Z',
          created_at: '2026-08-01T00:00:00Z',
        },
      ],
    });
    mockGetCircuitBreakerStatus.mockResolvedValue({
      gateway: 'razorpay',
      state: 'CLOSED',
      error_rate_rolling_1min: 0.0,
    });
  });

  test('feature-flag edit form renders the correct control type per scope_type', async () => {
    render(<SystemOpsPage />);

    // Switch to Feature Flags Tab
    fireEvent.click(screen.getByText('Feature Flags Admin'));

    await waitFor(() => {
      expect(screen.getByText('cart.dynamic_discounts')).toBeInTheDocument();
      expect(screen.getByText('warehouse.qc_required')).toBeInTheDocument();
      expect(screen.getByText('catalog.new_search')).toBeInTheDocument();
    });

    const editBtns = screen.getAllByText('Edit Rollout');

    // 1. GLOBAL Scope -> Checkbox toggle
    fireEvent.click(editBtns[0]);
    expect(screen.getByText('Enable Globally')).toBeInTheDocument();
    expect(screen.getByRole('checkbox')).toBeInTheDocument();

    // Close Modal
    fireEvent.click(screen.getByText('Cancel'));

    // 2. STORE Scope -> Text input for scope IDs
    fireEvent.click(editBtns[1]);
    expect(screen.getByPlaceholderText('e.g. store-mumbai-01, store-delhi-02')).toBeInTheDocument();

    // Close Modal
    fireEvent.click(screen.getByText('Cancel'));

    // 3. USER_PERCENTAGE Scope -> Range Slider
    fireEvent.click(editBtns[2]);
    expect(screen.getByText('Rollout Percentage')).toBeInTheDocument();
    expect(screen.getByRole('slider')).toBeInTheDocument();
  });

  test('DLQ discard action requires BOTH modal confirmation and step-up token before DELETE call fires', async () => {
    render(<SystemOpsPage />);

    await waitFor(() => {
      expect(screen.getByText('order.completed.dlq')).toBeInTheDocument();
    });

    // Inspect DLQ topic
    fireEvent.click(screen.getByText('order.completed.dlq'));

    await waitFor(() => {
      expect(screen.getByText('Offset #101')).toBeInTheDocument();
    });

    // Select message checkbox
    const checkbox = screen.getAllByRole('checkbox')[0];
    fireEvent.click(checkbox);

    // Click Discard Selected
    fireEvent.click(screen.getByText('Discard Selected (1)'));

    // Verify Warning modal copy is shown
    expect(
      screen.getByText(
        'This will permanently hide these failed events from the DLQ view. This does not undo whatever caused them to fail.'
      )
    ).toBeInTheDocument();

    // Ensure DELETE call has NOT fired yet
    expect(mockDiscardDlqMessages).not.toHaveBeenCalled();

    // Fill Step-Up token
    fireEvent.change(screen.getByPlaceholderText('Enter Step-Up Token'), {
      target: { value: 'stepup-secret-123' },
    });

    // Click Confirm Discard
    fireEvent.click(screen.getByText('Confirm Discard'));

    await waitFor(() => {
      expect(mockDiscardDlqMessages).toHaveBeenCalledWith(
        'order.completed.dlq',
        [101],
        'Unrecoverable dead-letter message',
        'stepup-secret-123'
      );
    });
  });
});

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ERPCredentialModal } from '../src/ERPCredentialModal';

describe('ERPCredentialModal Component Confirmation Test', () => {
  test('Requires confirmation checkbox before dismissal button is enabled', () => {
    const onConfirmedMock = jest.fn();

    render(
      <ERPCredentialModal
        connectionId="conn-100"
        integrationMode="AGENT_POLLED"
        webhookSecret="whsec_test_secret_123"
        agentApiKey="agent_key_test_token_456"
        connectorSetupNote="Test setup note"
        onConfirmed={onConfirmedMock}
      />
    );

    const closeBtn = screen.getByTestId('close-erp-credential-modal-btn');
    expect(closeBtn).toBeDisabled();

    // Check confirmation checkbox
    const checkbox = screen.getByTestId('confirm-download-checkbox');
    fireEvent.click(checkbox);

    expect(closeBtn).toBeEnabled();

    // Click close button
    fireEvent.click(closeBtn);
    expect(onConfirmedMock).toHaveBeenCalledTimes(1);
  });
});

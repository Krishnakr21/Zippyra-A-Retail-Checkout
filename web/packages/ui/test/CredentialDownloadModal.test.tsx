import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CredentialDownloadModal } from '../src/CredentialDownloadModal';

describe('CredentialDownloadModal Component', () => {
  const mockBundle = {
    certPem: '-----BEGIN CERTIFICATE-----\nTEST_CERT\n-----END CERTIFICATE-----',
    privateKeyPem: '-----BEGIN PRIVATE KEY-----\nTEST_KEY\n-----END PRIVATE KEY-----',
    rootCaPem: '-----BEGIN CERTIFICATE-----\nROOT_CA\n-----END CERTIFICATE-----',
    deviceJwt: 'mock.jwt.token',
  };

  test('Done button is disabled until confirmation checkbox is checked', () => {
    const onConfirmed = jest.fn();
    render(<CredentialDownloadModal bundle={mockBundle} onConfirmed={onConfirmed} />);

    const closeBtn = screen.getByTestId('close-credential-modal-btn');
    expect(closeBtn).toBeDisabled();

    const checkbox = screen.getByTestId('confirm-download-checkbox');
    fireEvent.click(checkbox);

    expect(closeBtn).toBeEnabled();
    fireEvent.click(closeBtn);
    expect(onConfirmed).toHaveBeenCalledTimes(1);
  });

  test('Download bundle triggers file download url creation', () => {
    const onConfirmed = jest.fn();
    const createObjectURLMock = jest.fn().mockReturnValue('blob:http://localhost/mock-url');
    const revokeObjectURLMock = jest.fn();

    global.URL.createObjectURL = createObjectURLMock;
    global.URL.revokeObjectURL = revokeObjectURLMock;

    render(<CredentialDownloadModal bundle={mockBundle} onConfirmed={onConfirmed} />);

    const downloadBtn = screen.getByTestId('download-bundle-btn');
    fireEvent.click(downloadBtn);

    expect(createObjectURLMock).toHaveBeenCalledTimes(1);
  });
});

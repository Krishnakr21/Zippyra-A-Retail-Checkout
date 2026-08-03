import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { OfferForm } from '../src/OfferForm';

describe('OfferForm Component Validation', () => {
  test('Rejects percent > 90 before firing onSubmit', async () => {
    const onSubmitMock = jest.fn();
    const onCancelMock = jest.fn();

    render(
      <OfferForm
        mode="create"
        scope="STORE"
        onSubmit={onSubmitMock}
        onCancel={onCancelMock}
      />
    );

    // Enter percent = 95
    const percentInput = screen.getByLabelText(/Discount Percent/i);
    fireEvent.change(percentInput, { target: { value: '95' } });

    const submitBtn = screen.getByTestId('submit-offer-btn');
    fireEvent.click(submitBtn);

    // Verify error message rendered and onSubmit NOT called
    await waitFor(() => {
      expect(screen.getByTestId('offer-form-error')).toBeInTheDocument();
      expect(screen.getByText('Percent off must be between 1% and 90%')).toBeInTheDocument();
    });

    expect(onSubmitMock).not.toHaveBeenCalled();
  });

  test('Rejects BOGO get_qty > buy_qty before firing onSubmit', async () => {
    const onSubmitMock = jest.fn();
    const onCancelMock = jest.fn();

    render(
      <OfferForm
        mode="create"
        scope="STORE"
        onSubmit={onSubmitMock}
        onCancel={onCancelMock}
      />
    );

    // Select BOGO type
    const typeSelect = screen.getByLabelText(/Offer Type/i);
    fireEvent.change(typeSelect, { target: { value: 'BOGO' } });

    // Set buy_qty = 1, get_qty = 2
    const buyInput = screen.getByLabelText(/Buy Quantity/i);
    const getInput = screen.getByLabelText(/Get Free Quantity/i);

    fireEvent.change(buyInput, { target: { value: '1' } });
    fireEvent.change(getInput, { target: { value: '2' } });

    const submitBtn = screen.getByTestId('submit-offer-btn');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByTestId('offer-form-error')).toBeInTheDocument();
      expect(screen.getByText('Get quantity cannot exceed Buy quantity')).toBeInTheDocument();
    });

    expect(onSubmitMock).not.toHaveBeenCalled();
  });
});

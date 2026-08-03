import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import StaffPage from '../app/dashboard/staff/page';

// Mock hooks
const mockGetStaffList = jest.fn();
const mockCreateStaff = jest.fn();
const mockUpdateStaff = jest.fn();
const mockDeactivateStaff = jest.fn();
const mockGetShiftHistory = jest.fn();

jest.mock('@zippyra/hooks', () => ({
  useStaff: () => ({
    getStaffList: mockGetStaffList,
    createStaff: mockCreateStaff,
    updateStaff: mockUpdateStaff,
    deactivateStaff: mockDeactivateStaff,
  }),
  useShiftHistory: () => ({
    getShiftHistory: mockGetShiftHistory,
  }),
}));

describe('StaffPage Component Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetStaffList.mockResolvedValue({
      staff: [
        {
          id: 'staff-1',
          name: 'Ramesh Kumar',
          phone: '+919876543210',
          role: 'CASHIER',
          is_active: true,
          has_pin_set: true,
          store_id: 'store-mumbai-01',
          created_at: '2026-08-01T00:00:00Z',
        },
        {
          id: 'staff-2',
          name: 'Suresh Patel',
          phone: '+919876543211',
          role: 'SECURITY',
          is_active: true,
          has_pin_set: false,
          store_id: 'store-mumbai-01',
          created_at: '2026-08-01T00:00:00Z',
        },
      ],
    });
    mockGetShiftHistory.mockResolvedValue({ shifts: [] });
  });

  test('renders has_pin_set indicator correctly for both true and false states', async () => {
    render(<StaffPage />);

    await waitFor(() => {
      expect(screen.getByText('Ramesh Kumar')).toBeInTheDocument();
      expect(screen.getByText('Suresh Patel')).toBeInTheDocument();
    });

    // Check PIN indicators
    expect(screen.getByText('🔑 PIN Set')).toBeInTheDocument();
    expect(screen.getByText('🔒 OTP Only')).toBeInTheDocument();
  });

  test('renders specific inline error on PHONE_ALREADY_STAFF conflict', async () => {
    mockCreateStaff.mockRejectedValue({
      code: 'PHONE_ALREADY_STAFF',
      status: 409,
      message: 'Staff phone already registered',
    });

    render(<StaffPage />);

    await waitFor(() => {
      expect(screen.getByText('+ Add Staff')).toBeInTheDocument();
    });

    // Open Add Modal
    fireEvent.click(screen.getByText('+ Add Staff'));

    // Fill form
    fireEvent.change(screen.getByPlaceholderText('e.g. Ramesh Kumar'), {
      target: { value: 'Duplicate User' },
    });
    fireEvent.change(screen.getByPlaceholderText('+919876543210'), {
      target: { value: '+919876543210' },
    });

    // Submit
    fireEvent.click(screen.getByText('Add Member'));

    await waitFor(() => {
      expect(
        screen.getByText('This number is already registered as staff somewhere on the platform')
      ).toBeInTheDocument();
    });
  });

  test('deactivate action requires ConfirmDialog confirmation before DELETE call fires', async () => {
    render(<StaffPage />);

    await waitFor(() => {
      expect(screen.getByText('Ramesh Kumar')).toBeInTheDocument();
    });

    // Click Deactivate on first staff member
    const deactivateBtns = screen.getAllByText('Deactivate');
    fireEvent.click(deactivateBtns[0]);

    // ConfirmDialog should be visible with specific warning copy
    expect(
      screen.getByText(
        'This will immediately log this staff member out everywhere and end any active shift. This cannot be undone without re-adding them.'
      )
    ).toBeInTheDocument();

    // Ensure deactivate API call has NOT fired yet
    expect(mockDeactivateStaff).not.toHaveBeenCalled();

    // Click Confirm button on modal
    fireEvent.click(screen.getByText('Deactivate Staff'));

    await waitFor(() => {
      expect(mockDeactivateStaff).toHaveBeenCalledWith('staff-1');
    });
  });
});

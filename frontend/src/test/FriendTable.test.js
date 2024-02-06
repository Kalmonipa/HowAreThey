import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import 'mutationobserver-shim';
import fetchMock from 'jest-fetch-mock';
import FriendTable from '../components/FriendTable';

jest.mock('../components/Modal', () => ({ show, children, onClose }) => show ? (
  <div role="dialog">
    {children}
    <button onClick={onClose}>Close Modal</button>
  </div>
) : null);

beforeEach(() => {
  fetchMock.resetMocks();
});

test('renders without crashing and fetches friends data', async () => {
  const mockFriends = [
    { ID: "1", Name: 'Alice', LastContacted: '01/01/2021', Notes: 'Loves coding' },
    { ID: "2", Name: 'Bob', LastContacted: '01/01/2021', Notes: 'Enjoys hiking' },
  ];

  fetchMock.mockResponseOnce(JSON.stringify(mockFriends));

  render(<FriendTable friends={[]} filterText="" />);

  await waitFor(() => {
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });
});

test('filters friends correctly', async () => {
  const mockFriends = [
    { ID: "1", Name: 'Alice', LastContacted: '01/01/2021', Notes: 'Loves coding' },
    { ID: "2", Name: 'Bob', LastContacted: '01/01/2021', Notes: 'Enjoys hiking' },
  ];

  fetchMock.mockResponseOnce(JSON.stringify(mockFriends));

  render(<FriendTable friends={mockFriends} filterText="Alice" />);

  await waitFor(() => {
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.queryByText('Bob')).toBeNull();
  });
});

import React from 'react';
import {render, screen} from '@testing-library/react';
import FriendTable from '../components/FriendTable';

describe('FriendTable Component Tests', () => {
  const mockFriends = [
    {Name: 'Alice', LastContacted: '01/01/2021', Notes: 'Loves coding'},
    {Name: 'Bob', LastContacted: '01/01/2021', Notes: 'Enjoys hiking'},
  ];

  it('renders without crashing', () => {
    render(<FriendTable friends={mockFriends} filterText="" />);
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('filters friends correctly', () => {
    render(<FriendTable friends={mockFriends} filterText="Alice" />);
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.queryByText('Bob')).toBeNull();
  });

  it('renders the correct number of friends', () => {
    render(<FriendTable friends={mockFriends} filterText="" />);
    const rows = screen.getAllByRole('row');
    // Expect one extra for the header row
    expect(rows.length).toBe(mockFriends.length + 1);
  });

  it('renders table headers correctly', () => {
    render(<FriendTable friends={mockFriends} filterText="" />);
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Last Contacted')).toBeInTheDocument();
    expect(screen.getByText('Notes')).toBeInTheDocument();
  });
});


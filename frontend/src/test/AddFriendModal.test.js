import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import AddFriendModal from '../components/AddFriendModal'; // Adjust the import path as needed

test('renders correctly when show is true', () => {
  render(<AddFriendModal show={true} onClose={() => {}} onSaved={() => {}} />);
  expect(screen.getByText(/Add a new friend/i)).toBeInTheDocument();
});

test('does not render when show is false', () => {
  render(<AddFriendModal show={false} onClose={() => {}} onSaved={() => {}} />);
  expect(screen.queryByText(/Add a new friend/i)).not.toBeInTheDocument();
});

test('displays alert if name input is empty on save', () => {
  window.alert = jest.fn();

  render(<AddFriendModal show={true} onClose={() => {}} onSaved={() => {}} />);

  fireEvent.click(screen.getByText(/Save/i));
  expect(window.alert).toHaveBeenCalledWith('Name field is empty');
});

test('displays alert if no date is selected on save', () => {
    window.alert = jest.fn();

    render(<AddFriendModal show={true} onClose={() => {}} onSaved={() => {}} />);

    fireEvent.change(screen.getByPlaceholderText(/Name.../i), { target: { value: 'John Doe' } });
    fireEvent.click(screen.getByText(/Save/i));
    expect(window.alert).toHaveBeenCalledWith('No date selected. Enter approximate date if unknown');
});

test('successfully submits form with valid data', async () => {
  window.alert = jest.fn();
  fetch.mockResponseOnce(JSON.stringify({ message: "Jane Doe added successfully" }), { status: 201 });

  render(<AddFriendModal show={true} onClose={() => {}} onSaved={() => {}} />);

  // Fill out the form...
  fireEvent.change(screen.getByPlaceholderText(/Name.../i), { target: { value: 'Jane Doe' } });
  fireEvent.change(screen.getByPlaceholderText(/Select a date.../i), { target: { value: '06/06/2023' } });


  fireEvent.click(screen.getByText(/Save/i));


    expect(fetch).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
        Name: 'Jane Doe',
        LastContacted: '06/06/2023',
        Notes: ''
        })
    }));
});

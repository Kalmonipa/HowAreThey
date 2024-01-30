import React from 'react';
import {render, screen, fireEvent} from '@testing-library/react';
import SearchBar from './SearchBar'; // Adjust the import path as necessary

describe('SearchBar Component', () => {
  test('renders search input', () => {
    render(<SearchBar filterText="" onFilterTextChange={() => {}} />);
    const inputElement = screen.getByPlaceholderText(/search.../i);
    expect(inputElement).toBeInTheDocument();
  });

  test('displays the correct filter text', () => {
    const filterText = 'Test';
    render(<SearchBar filterText={filterText} onFilterTextChange={() => {}} />);
    const inputElement = screen.getByPlaceholderText(/search.../i);
    expect(inputElement.value).toBe(filterText);
  });

  test('calls onFilterTextChange on input change', () => {
    const handleFilterTextChange = jest.fn();
    render(<
      SearchBar filterText="" onFilterTextChange={handleFilterTextChange}
    />);

    const inputElement = screen.getByPlaceholderText(/search.../i);
    fireEvent.change(inputElement, { target: { value: 'New Value' } });

    expect(handleFilterTextChange).toHaveBeenCalledTimes(1);
    expect(handleFilterTextChange).toHaveBeenCalledWith('New Value');
  });
});

import React from 'react';
import {render, screen, waitFor} from '@testing-library/react';
import fetchMock from 'jest-fetch-mock';
import App from '../src/App';
import 'mutationobserver-shim';

beforeEach(() => {
  fetchMock.resetMocks();
});

test('fetches friends on mount', async () => {
  const mockFriends = [{ ID: "1", Name: 'John Doe', LastContacted: "06/06/2023", Notes: "" }];
  fetchMock.mockResponse(JSON.stringify(mockFriends));

  render(<App />);

  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  screen.getByText('John Doe');
});

import React, { useState, useEffect } from 'react';
import RandomFriendButton from './components/RandomFriendButton';
import SearchBar from './components/SearchBar';
import PageHeader from './components/PageHeader';
import FriendTable from './components/FriendTable';

import './css/FriendTable.css';
import './css/App.css';


function FilterableFriendsTable({ friends }) {
  const [filterText, setFilterText] = useState('');

  return (
    <div className='body'>
      <PageHeader />
      <SearchBar
        filterText={filterText}
        onFilterTextChange={setFilterText}
      />
      <FriendTable
        friends={friends}
        filterText={filterText}
      />
    </div>
  );
}

export default function App() {
  const [friends, setFriends] = useState([]);

  useEffect(() => {
    fetch('http://localhost:8080/friends')
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        return response.json();
      })
      .then(data => setFriends(data))
      .catch(error => {
        console.error('Error fetching data:', error);
      });
  }, []);

  const handleRandomFriend = () => {
    fetch('http://localhost:8080/friends/random')
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok ' + response.statusText);
        }
        return response.json();
      })
      .catch(error => {
        console.error('There has been a problem with your fetch operation:', error);
      });
  };

  return (
    <div>
      <FilterableFriendsTable friends={friends} />
      <RandomFriendButton onRandomFriendSelect={handleRandomFriend} />
    </div>
  );
}

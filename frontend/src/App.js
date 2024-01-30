import React, {useState, useEffect} from 'react';
import PropTypes from 'prop-types';

import EditButton from './components/EditButton';
import FriendTable from './components/FriendTable';
import PageHeader from './components/PageHeader';
import SearchBar from './components/SearchBar';
import RandomFriendButton from './components/RandomFriendButton';


import './css/ActivityBar.css'
import './css/FriendTable.css';
import './css/App.css';


function FilterableFriendsTable({friends}) {
  const [filterText, setFilterText] = useState('');
  const [isEditable, setIsEditable] = useState(false);

  const toggleEdit = () => {
    setIsEditable(!isEditable);
  };

  const handleRandomFriend = () => {
    fetch('http://localhost:8080/friends/random')
        .then((response) => {
          if (!response.ok) {
            throw new Error('Network response was not ok ',
                + response.statusText);
          }
          return response.json();
        })
        .catch((error) => {
          console.error('There has been a problem with your fetch operation:',
              error);
        });
  };

  return (
    <div className='body'>
      <PageHeader />
      <div className="activities-bar">
        <SearchBar
          filterText={filterText}
          onFilterTextChange={setFilterText}
        />
        <div className="button-group">
          <EditButton isEditable={isEditable} onClick={toggleEdit} />
          <RandomFriendButton onRandomFriendSelect={handleRandomFriend} />
        </div>
      </div>
      <FriendTable
        friends={friends}
        filterText={filterText}
        isEditable={isEditable}
      />
    </div>
  );
}
FilterableFriendsTable.propTypes = {
  friends: PropTypes.array.isRequired,
};

export default function App() {
  const [friends, setFriends] = useState([]);

  useEffect(() => {
    fetch('http://localhost:8080/friends')
        .then((response) => {
          if (!response.ok) {
            throw new Error('Network response was not ok');
          }
          return response.json();
        })
        .then((data) => setFriends(data))
        .catch((error) => {
          console.error('Error fetching data:', error);
        });
  }, []);

  return (
    <div>
      <FilterableFriendsTable friends={friends} />
    </div>
  );
}

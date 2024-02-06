import React, {useState, useEffect, useCallback} from 'react';
import PropTypes from 'prop-types';

import AddFriendButton from './components/AddFriendButton';
import FriendTable from './components/FriendTable';
import PageHeader from './components/PageHeader';
import SearchBar from './components/SearchBar';
import SettingsButton from './components/SettingsButton';
import RandomFriendButton from './components/RandomFriendButton';

import './css/ActivityBar.css'
import './css/FriendTable.css';
import './css/App.css';


function FilterableFriendsTable({friends, setFriends}) {
  const [filterText, setFilterText] = useState('');
  const [isEditable, setIsEditable] = useState(false);


  const toggleEdit = () => {
    setIsEditable(!isEditable);
  };

  const fetchFriends = useCallback(() => {
    fetch('http://localhost:8080/friends')
      .then(response => response.json())
      .then(data => setFriends(data))
      .catch(error => console.error('Error fetching data:', error));
  }, []);

  useEffect(() => {
    fetchFriends();
  }, [fetchFriends]);

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
          <AddFriendButton fetchFriends={fetchFriends} />
          <RandomFriendButton onRandomFriendSelect={handleRandomFriend} />
          <SettingsButton isEditable={isEditable} onClick={toggleEdit} />
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
      <FilterableFriendsTable friends={friends} setFriends={setFriends} />
    </div>
  );

}

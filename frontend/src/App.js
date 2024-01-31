import React, {useState, useEffect} from 'react';
import PropTypes from 'prop-types';

import AddFriendButton from './components/AddFriendButton';
import AddFriendModal from './components/AddFriendModal';
import FriendTable from './components/FriendTable';
import PageHeader from './components/PageHeader';
import SearchBar from './components/SearchBar';
import SettingsButton from './components/SettingsButton';
import RandomFriendButton from './components/RandomFriendButton';


import './css/ActivityBar.css'
import './css/FriendTable.css';
import './css/App.css';


function FilterableFriendsTable({friends, onAddFriendClick}) {
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
          <AddFriendButton addFriendSelect={onAddFriendClick} />
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
  onAddFriendClick: PropTypes.func.isRequired,
};

export default function App() {
  const [friends, setFriends] = useState([]);
  const [showModal, setShowModal] = useState(false);

  const handleAddFriendClick = () => {
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
  };

  const fetchFriends = () => {
    fetch('http://localhost:8080/friends')
      .then(response => response.json())
      .then(data => setFriends(data))
      .catch(error => console.error('Error fetching data:', error));
  };

  const handleSave = () => {
    fetchFriends();
    setShowModal(false);
  };

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
      <FilterableFriendsTable
      friends={friends}
      onAddFriendClick={handleAddFriendClick}
      />
      <AddFriendModal
        show={showModal}
        onClose={handleCloseModal}
        onSaved={handleSave}>
      </AddFriendModal>
    </div>
  );
}

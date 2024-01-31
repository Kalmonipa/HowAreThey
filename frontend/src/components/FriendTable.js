import React, {useState, useEffect} from 'react';
import PropTypes from 'prop-types';
import Modal from './Modal';

import '../css/FriendTable.css';
import '../css/EditBoxes.css';

function FriendTable({ friends, filterText }) {
  const [editableRowId, setEditableRowId] = useState(null);
  const [friendsData, setFriendsData] = useState(friends);
  const [showModal, setShowModal] = useState(false);
  const [modalContent, setModalContent] = useState('');

  const fetchFriendsData = async () => {
    const response = await fetch('http://localhost:8080/friends');
    const data = await response.json();
    setFriendsData(data);
  };

  useEffect(() => {
    fetchFriendsData();
  }, []);

  const exitEditMode = () => {
    setEditableRowId(null);
  };

  const rows = friendsData.map((friend) => {
    if (friend.Name.toLowerCase().includes(filterText.toLowerCase())) {
      return (
        <FriendRow
          friend={friend}
          key={friend.ID}
          editable={friend.ID === editableRowId}
          onRowClick={() => setEditableRowId(friend.ID)}
          onExitEditMode={exitEditMode}
          fetchFriendsData={fetchFriendsData}
          setShowModal={setShowModal}
          setModalContent={setModalContent}
        />
      );
    }
    return null;
  });

  return (
    <div>
    <Modal show={showModal} onClose={() => setShowModal(false)}>
      <p>{modalContent}</p>
    </Modal>
      <table className="friend-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Last Contacted</th>
            <th>Notes</th>
          </tr>
        </thead>
      <tbody>{rows}</tbody>
      </table>
    </div>
  );
}

FriendTable.propTypes = {
  friends: PropTypes.arrayOf(PropTypes.object).isRequired,
  filterText: PropTypes.string.isRequired,
};

function FriendRow({
  friend,
  editable,
  onRowClick,
  onExitEditMode,
  fetchFriendsData,
  setModalContent,
  setShowModal,
 }) {
  const [updatedFriend, setUpdatedFriend] = useState({ ...friend });

  const handleInputChange = (field, value) => {
    setUpdatedFriend(prev => ({ ...prev, [field]: value }));
  };

  const handleKeyPress = async (event) => {
    if (event.key === 'Enter') {
      await handleSave();
      onExitEditMode();
    }
  };

  const handleClickOutside = async (event) => {
    if (event.target.closest('.friend-table-row') !== null) {
      return;
    }
    await handleSave();
    onExitEditMode();
  };

  useEffect(() => {
    if (editable) {
      // Add event listener when the row is editable
      document.addEventListener('click', handleClickOutside);
    }
    return () => {
      // Clean up the event listener when the component unmounts or becomes non-editable
      document.removeEventListener('click', handleClickOutside);
    };
  }, [editable, onExitEditMode]);

  const isValidDate = (dateString) => {
    const regex = /^(\d{2})\/(\d{2})\/(\d{4})$/;
    if (dateString.match(regex) === null) {
      return false;
    }
    const [day, month, year] = dateString.split('/').map(Number);
    const date = new Date(year, month - 1, day);
    return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day;
  };


  const handleSave = async () => {
    if (!isValidDate(updatedFriend.LastContacted)) {
      setModalContent('Invalid date format. Please use DD/MM/YYYY.');
      setShowModal(true);
      return;
    }
    try {
      const response = await fetch(`http://localhost:8080/friends/${friend.ID}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedFriend),
      });
      if (response.ok) {
        await fetchFriendsData();
      }
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      // Optionally, update the friends list in the parent component here
      onExitEditMode();
    } catch (error) {
      console.error('Failed to update friend:', error);
    }
  };

  const renderCell = (content, field, isEditable) => {
    return isEditable ? (
      <td>
        <input
          type="text"
          value={updatedFriend[field]}
          className="editable-input"
          onChange={(e) => handleInputChange(field, e.target.value)}
          onKeyDown={handleKeyPress}
        />
      </td>
    ) : (
      <td>{content}</td>
    );
  };

  return (
    <tr className="friend-table-row" onClick={onRowClick}>
      {renderCell(friend.ID, 'ID', false)}
      {renderCell(friend.Name, 'Name', editable)}
      {renderCell(friend.LastContacted, 'LastContacted', editable)}
      {renderCell(friend.Notes, 'Notes', editable)}
    </tr>
  );
}


FriendRow.propTypes = {
  friend: PropTypes.shape({
    Name: PropTypes.string,
    LastContacted: PropTypes.string,
    Notes: PropTypes.string,
  }).isRequired,
  onExitEditMode: PropTypes.func.isRequired,
  setShowModal: PropTypes.func.isRequired,
  setModalContent: PropTypes.func.isRequired,
};

export default FriendTable;

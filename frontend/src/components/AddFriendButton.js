import React, {useState} from 'react';
import PropTypes from 'prop-types';
import AddFriendModal from './AddFriendModal';

import 'react-datepicker/dist/react-datepicker.css';
import '../css/AddFriendButton.css';


function AddFriendButton({ fetchFriends }) {
  const [showModal, setShowModal] = useState(false);

  const handleAddFriendClick = () => {
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
  };

  const handleSave = () => {
    setShowModal(false);
    fetchFriends();
  };

  return (
    <div>
      <button onClick={handleAddFriendClick} className="add-friend-button">
        Add Friend
      </button>
      <AddFriendModal
        show={showModal}
        onClose={handleCloseModal}
        onSaved={handleSave}>
      </AddFriendModal>
    </div>
  );
}

AddFriendButton.propTypes = {
  fetchFriends: PropTypes.func.isRequired,
};

export default AddFriendButton;

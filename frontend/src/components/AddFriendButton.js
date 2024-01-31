import React from 'react';
import 'react-datepicker/dist/react-datepicker.css';
import PropTypes from 'prop-types';

import 'react-datepicker/dist/react-datepicker.css';
import '../css/AddFriendButton.css';

function AddFriendButton({addFriendSelect}) {
  return (
    <button onClick={addFriendSelect} className="add-friend-button">
      Add Friend
    </button>
  );
}

AddFriendButton.propTypes = {
  addFriendSelect: PropTypes.func.isRequired,
};

export default AddFriendButton;

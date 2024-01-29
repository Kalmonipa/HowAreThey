import React from 'react';
import PropTypes from 'prop-types';

import '../css/RandomFriendButton.css';

function RandomFriendButton({onRandomFriendSelect}) {
  return (
    <button onClick={onRandomFriendSelect} className="random-friend-button">
      Pick a friend to contact
    </button>
  );
}

RandomFriendButton.propTypes = {
  onRandomFriendSelect: PropTypes.func.isRequired,
};

export default RandomFriendButton;

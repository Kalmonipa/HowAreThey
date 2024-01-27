import React from 'react';

function RandomFriendButton() {
    function handleClick() {
        alert('You clicked me!');
    }
    return (
      <button onClick={handleClick}>
        Pick a friend to contact
      </button>
    );
  }

export default RandomFriendButton;

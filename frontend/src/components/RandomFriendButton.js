import React from 'react';
import '../css/RandomFriendButton.css'

function RandomFriendButton({ onRandomFriendSelect }) {
    return (
        <button onClick={onRandomFriendSelect} className='random-friend-button'>
            Pick a friend to contact
        </button>
    );
}

export default RandomFriendButton;

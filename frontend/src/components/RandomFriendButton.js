import React from 'react';

function RandomFriendButton({ onRandomFriendSelect }) {
    return (
        <button onClick={onRandomFriendSelect}>
            Pick a friend to contact
        </button>
    );
}

export default RandomFriendButton;
